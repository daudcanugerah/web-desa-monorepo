#!/usr/bin/env python3
"""
Seed sample Metabase dashboards for Desa Palasari.

Creates questions (cards) + dashboards in the target collection from two
uploaded CSV datasets (kependudukan_umur, indikator_desa), then optionally
registers each dashboard into the desa API infographic table.

Idempotent: looks up existing cards/dashboards by name before creating.
Tables are discovered dynamically by name prefix, so re-running after a
re-upload still works.

Usage:
    export MB_URL=http://bi.desapalasari.my.id
    export MB_USER=daudcanugerah@gmail.com
    export MB_PASSWORD='...'
    python3 seed_metabase.py --dry-run
    python3 seed_metabase.py
    python3 seed_metabase.py --only Pendidikan

NOTE: password is read from the environment, never written to disk.
"""
import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

MB_URL = os.environ.get("MB_URL", "http://bi.desapalasari.my.id").rstrip("/")
MB_USER = os.environ.get("MB_USER", "")
MB_PASSWORD = os.environ.get("MB_PASSWORD", "")
MB_DB_ID = int(os.environ.get("MB_DB_ID", "2"))
MB_COLLECTION_ID = int(os.environ.get("MB_COLLECTION_ID", "5"))
# Existing "Penduduk" dashboard that gets the age-pyramid card appended.
MB_PENDUDUK_DASHBOARD_ID = int(os.environ.get("MB_PENDUDUK_DASHBOARD_ID", "2"))

DRY_RUN = False
SESSION = None


# ---------------------------------------------------------------- HTTP helpers
def api(method, path, data=None, raw=None, content_type="application/json"):
    url = MB_URL + path
    headers = {}
    if SESSION:
        headers["X-Metabase-Session"] = SESSION
    if raw is not None:
        headers["Content-Type"] = content_type
        body = raw
    elif data is not None:
        headers["Content-Type"] = "application/json"
        body = json.dumps(data).encode()
    else:
        body = None
    req = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            raw_body = resp.read()
            return json.loads(raw_body) if raw_body else {}
    except urllib.error.HTTPError as e:
        detail = e.read().decode()
        raise RuntimeError(f"{method} {path} -> {e.code}: {detail[:500]}")


def login():
    global SESSION
    res = api("POST", "/api/session", {"username": MB_USER, "password": MB_PASSWORD})
    SESSION = res["id"]
    me = api("GET", "/api/user/current")
    if not me.get("is_superuser"):
        raise RuntimeError("user is not a superuser")
    print(f"logged in as {me.get('email')} (id {me.get('id')})")


def find_table(prefix):
    """Return the most recent table id whose name starts with prefix."""
    meta = api("GET", f"/api/database/{MB_DB_ID}/metadata")
    matches = [t for t in meta.get("tables", []) if t.get("name", "").startswith(prefix)]
    if not matches:
        raise RuntimeError(f"no table found with prefix '{prefix}' — upload the CSV first")
    # uploaded table names end in a timestamp → newest sorts last lexicographically
    matches.sort(key=lambda t: t["name"])
    return matches[-1]["name"]


def find_card_by_name(name):
    items = api("GET", f"/api/collection/{MB_COLLECTION_ID}/items?models=card")
    for it in items.get("data", []):
        if it.get("name") == name and not it.get("archived"):
            return it
    return None


def find_dashboard_by_name(name):
    dashes = api("GET", "/api/dashboard")
    root = dashes if isinstance(dashes, list) else dashes.get("data", [])
    for d in root:
        if d.get("name") == name and not d.get("archived"):
            return d
    # fall back to collection items scan
    items = api("GET", f"/api/collection/{MB_COLLECTION_ID}/items?models=dashboard")
    for it in items.get("data", []):
        if it.get("name") == name:
            return it
    return None


# ------------------------------------------------------------------ definitions
# Each card: key -> (name, display, sql, description, viz)
def q_age():
    return ("Piramida Penduduk 2021", "bar",
            "SELECT kelompok_umur, jumlah FROM {umur} ORDER BY urutan",
            "Distribusi penduduk Desa Palasari menurut kelompok umur (2021).",
            {"graph.dimensions": ["kelompok_umur"], "graph.metrics": ["jumlah"],
             "graph.x_axis.title_text": "Kelompok Umur", "graph.y_axis.title_text": "Jiwa"})


def q(kategori, kelompok, name, display, desc, xlabel=None, ylabel="Jumlah"):
    where = f"WHERE kategori = '{kategori}'"
    if kelompok == "__empty":
        where += " AND kelompok IS NULL"
    elif kelompok:
        where += f" AND kelompok = '{kelompok}'"
    sql = (f"SELECT indikator, nilai FROM {{indikator}} {where} "
           f"ORDER BY nilai DESC")
    viz = {"graph.dimensions": ["indikator"], "graph.metrics": ["nilai"]}
    if xlabel:
        viz["graph.x_axis.title_text"] = xlabel
    if display in ("bar", "line"):
        viz["graph.y_axis.title_text"] = ylabel
    return (name, display, sql, desc, viz)


def q_filter(indikator_like, name, display, desc, ylabel="Jumlah"):
    sql = (f"SELECT indikator, nilai FROM {{indikator}} "
           f"WHERE indikator LIKE '{indikator_like}' ORDER BY nilai DESC")
    return (name, display, sql, desc,
            {"graph.dimensions": ["indikator"], "graph.metrics": ["nilai"],
             "graph.y_axis.title_text": ylabel})


CARDS = {
    "age":        lambda: q_age(),
    "pendidikan_attain": lambda: q("Pendidikan", "attainment", "Tingkat Pendidikan Penduduk", "bar",
                                   "Jumlah penduduk menurut jenjang pendidikan tertinggi (2021).", "Jenjang", "Jiwa"),
    "pendidikan_murid":  lambda: q("Pendidikan", "murid", "Jumlah Murid per Jenjang", "bar",
                                   "Jumlah murid SD, SMP, dan SMA/SMK di Desa Palasari (2021).", "Jenjang", "Siswa"),
    "pendidikan_sarana": lambda: q("Pendidikan", "sarana", "Sarana Pendidikan", "bar",
                                   "Jumlah sarana pendidikan di Desa Palasari (2021).", "Jenis", "Unit"),
    "pendidikan_guru":   lambda: q("Pendidikan", "guru", "Jumlah Guru per Jenjang", "bar",
                                   "Jumlah guru SD dan SMP di Desa Palasari (2021).", "Jenjang", "Guru"),
    "pekerjaan":         lambda: q("Pekerjaan", "", "Komposisi Pekerjaan Penduduk", "pie",
                                   "Komposisi mata pencaharian penduduk Desa Palasari (2021)."),
    "kesehatan_sarana":  lambda: q("Kesehatan", "__empty", "Sarana Kesehatan", "bar",
                                   "Jumlah sarana dan tenaga kesehatan di Desa Palasari (2021).", "Jenis", "Unit/Orang"),
    "kesehatan_disabilitas": lambda: q("Kesehatan", "disabilitas", "Penyandang Disabilitas", "bar",
                                       "Jumlah penyandang disabilitas menurut jenisnya (2021).", "Jenis", "Jiwa"),
    "kb":                lambda: q("KB", "akseptor", "Akseptor KB", "bar",
                                   "Jumlah akseptor KB menurut jenis kontrasepsi (2021).", "Alat Kontrasepsi", "Akseptor"),
    "pertanian_lahan":   lambda: q_filter("Lahan%", "Penggunaan Lahan Pertanian", "bar",
                                          "Luas penggunaan lahan pertanian dan non-pertanian (Ha, 2021).", "Ha"),
    "padi_produksi":     lambda: q_filter("Padi - Produksi", "Produksi Padi 2021", "scalar",
                                          "Produksi padi Desa Palasari tahun 2021 (Ton)."),
    "agama_pemeluk":     lambda: q_filter("Penduduk %", "Pemeluk Agama", "pie",
                                          "Komposisi penduduk menurut agama (2021)."),
    "agama_ibadah":      lambda: q_filter("%%", "Sarana Ibadah", "bar",
                                          "Jumlah sarana ibadah di Desa Palasari (2021).", "Unit"),
    "pemerintahan":      lambda: q("Pemerintahan", "", "Pemerintahan Desa", "table",
                                   "Data kelembagaan dan pemerintahan Desa Palasari (2021)."),
    "utilitas":          lambda: q("Utilitas", "", "Utilitas Rumah Tangga", "bar",
                                   "Akses rumah tangga terhadap listrik dan air bersih (2021).", "Jenis", "Keluarga/Persen"),
    "perdagangan":       lambda: q("Perdagangan", "", "Sarana Perdagangan", "bar",
                                   "Jumlah sarana perdagangan di Desa Palasari (2021).", "Jenis", "Unit"),
    "transportasi":      lambda: q("Transportasi", "", "Sarana Transportasi & Telekomunikasi", "bar",
                                   "Jumlah sarana transportasi dan telekomunikasi (2021).", "Jenis", "Unit"),
    "pbb":               lambda: q_filter("PBB - %", "PBB: Target vs Realisasi", "bar",
                                          "Perbandingan target dan realisasi PBB Desa Palasari (2021).", "Rupiah"),
}

# Fix the two special filters that need custom SQL.
def make_agama_ibadah():
    sql = ("SELECT indikator, nilai FROM {indikator} "
           "WHERE kategori = 'Agama' AND indikator IN ('Masjid','Mushola / Langgar') ORDER BY nilai DESC")
    return ("Sarana Ibadah", "bar", sql,
            "Jumlah masjid dan mushola di Desa Palasari (2021).",
            {"graph.dimensions": ["indikator"], "graph.metrics": ["nilai"],
             "graph.y_axis.title_text": "Unit"})


def make_agama_pemeluk():
    sql = ("SELECT indikator, nilai FROM {indikator} "
           "WHERE kategori = 'Agama' AND indikator LIKE 'Penduduk %' ORDER BY nilai DESC")
    return ("Pemeluk Agama", "pie", sql,
            "Komposisi penduduk menurut agama (2021).",
            {"pie.dimension": "indikator", "pie.metric": "nilai"})


CARDS["agama_ibadah"] = make_agama_ibadah
CARDS["agama_pemeluk"] = make_agama_pemeluk

DASHBOARDS = {
    "Pendidikan": ["pendidikan_attain", "pendidikan_murid", "pendidikan_sarana", "pendidikan_guru"],
    "Pekerjaan": ["pekerjaan"],
    "Kesehatan": ["kesehatan_sarana", "kesehatan_disabilitas", "kb"],
    "Pertanian": ["pertanian_lahan", "padi_produksi"],
    "Keagamaan": ["agama_pemeluk", "agama_ibadah"],
    "Pemerintahan": ["pemerintahan", "utilitas"],
    "Ekonomi & Infrastruktur": ["perdagangan", "transportasi", "pbb"],
}


def make_card(spec, umur_table, indikator_table):
    name, display, sql, desc, viz = spec
    sql = sql.replace("{umur}", umur_table).replace("{indikator}", indikator_table)
    existing = find_card_by_name(name)
    if existing:
        print(f"    card exists: {name} (id {existing['id']})")
        return existing["id"]
    if DRY_RUN:
        print(f"    [dry-run] create card: {name} [{display}]")
        return None
    payload = {
        "name": name,
        "description": desc,
        "collection_id": MB_COLLECTION_ID,
        "dataset_query": {"database": MB_DB_ID, "type": "native",
                          "native": {"query": sql, "template-tags": {}}},
        "display": display,
        "visualization_settings": viz,
    }
    res = api("POST", "/api/card", payload)
    print(f"    card created: {name} (id {res['id']})")
    return res["id"]


def layout(card_ids):
    """2-per-row grid, matching the existing Penduduk dashboard (12 wide, 7 tall)."""
    out = []
    for i, cid in enumerate(card_ids):
        out.append({
            "id": -(i + 1),  # negative temp id for new dashcards
            "card_id": cid,
            "row": (i // 2) * 7,
            "col": (i % 2) * 12,
            "size_x": 12,
            "size_y": 7,
        })
    return out


def ensure_dashboard(name, card_ids):
    existing = find_dashboard_by_name(name)
    if existing:
        print(f"  dashboard exists: {name} (id {existing['id']}) — leaving untouched")
        return existing["id"]
    if DRY_RUN:
        print(f"  [dry-run] create dashboard: {name} with {len(card_ids)} cards")
        return None
    d = api("POST", "/api/dashboard", {"name": name, "collection_id": MB_COLLECTION_ID})
    did = d["id"]
    print(f"  dashboard created: {name} (id {did})")
    # Static embedding is opt-in per object; without this the embed page
    # returns "Embedding is not enabled for this object."
    api("PUT", f"/api/dashboard/{did}", {"enable_embedding": True, "embedding_params": {}})
    api("PUT", f"/api/dashboard/{did}/cards", {"cards": layout(card_ids)})
    print(f"    placed {len(card_ids)} cards + enabled static embedding")
    return did


def append_age_pyramid(card_id):
    """Append the age-pyramid card to the existing Penduduk dashboard."""
    if card_id is None:
        print("  [dry-run] append age pyramid to Penduduk dashboard")
        return
    dash = api("GET", f"/api/dashboard/{MB_PENDUDUK_DASHBOARD_ID}")
    existing = dash.get("dashcards", [])
    if any(dc.get("card_id") == card_id for dc in existing):
        print("  age pyramid already on Penduduk dashboard")
        return
    cards = [{"id": dc["id"], "card_id": dc["card_id"], "row": dc["row"], "col": dc["col"],
              "size_x": dc["size_x"], "size_y": dc["size_y"]} for dc in existing]
    bottom = max((c["row"] + c["size_y"] for c in cards), default=0)
    cards.append({"id": -(len(cards) + 1), "card_id": card_id,
                  "row": bottom, "col": 0, "size_x": 24, "size_y": 7})
    api("PUT", f"/api/dashboard/{MB_PENDUDUK_DASHBOARD_ID}/cards", {"cards": cards})
    print(f"  appended age pyramid to Penduduk dashboard (id {MB_PENDUDUK_DASHBOARD_ID})")


def main():
    global DRY_RUN
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--only", help="create a single dashboard by name")
    args = ap.parse_args()
    DRY_RUN = args.dry_run

    if not MB_USER or not MB_PASSWORD:
        print("MB_USER and MB_PASSWORD env vars are required", file=sys.stderr)
        sys.exit(1)

    login()
    umur_table = find_table("kependudukan_umur")
    indikator_table = find_table("indikator_desa")
    print(f"source tables: {umur_table} / {indikator_table}")

    # 1) age pyramid card + append to existing Penduduk dashboard
    print("\n[age pyramid]")
    age_id = make_card(CARDS["age"](), umur_table, indikator_table)
    append_age_pyramid(age_id)

    # 2) remaining dashboards
    for dash_name, keys in DASHBOARDS.items():
        if args.only and args.only != dash_name:
            continue
        print(f"\n[{dash_name}]")
        ids = [make_card(CARDS[k](), umur_table, indikator_table) for k in keys]
        ensure_dashboard(dash_name, [i for i in ids if i])

    print("\ndone.")


if __name__ == "__main__":
    main()