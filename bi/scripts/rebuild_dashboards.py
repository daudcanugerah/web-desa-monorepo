#!/usr/bin/env python3
"""
Rebuild sample Metabase dashboards for Desa Palasari as *data summaries*:
each dashboard has a summary section (latest KPIs / distribution) plus
historical (line) charts where the BPS data provides >= 2 years.

Pattern mirrors the user's own "Penduduk" dashboard (summary + YoY lines).

Data sources (uploaded to Metabase DB 2, collection 5):
  - indikator_derived.csv    latest-value derived metrics (%, ratios, scalars)
  - indikator_historis.csv   tidy year series: dashboard,indikator,tahun,nilai,satuan
  - kependudukan_umur.csv    age pyramid (kept on Penduduk dashboard)

Clean rebuild (name-based, robust to id churn):
  1. Delete our dashboards + cards in collection 5 (never the user's cards
     42/43, never the Penduduk dashboard itself).
  2. Remove stale infographic rows in the desa API.
  3. Upload CSVs, create summary + historical cards, create dashboards,
     enable static embedding.
  4. Register dashboards in the desa API infographic table.

Usage:
    export MB_URL=... MB_USER=... MB_PASSWORD=...
    python3 rebuild_dashboards.py --dry-run
    python3 rebuild_dashboards.py
"""
import argparse
import json
import mimetypes
import os
import sys
import uuid
import urllib.error
import urllib.request

MB_URL = os.environ.get("MB_URL", "http://bi.desapalasari.my.id").rstrip("/")
MB_USER = os.environ.get("MB_USER", "")
MB_PASSWORD = os.environ.get("MB_PASSWORD", "")
MB_DB_ID = int(os.environ.get("MB_DB_ID", "2"))
MB_COLLECTION_ID = int(os.environ.get("MB_COLLECTION_ID", "5"))
PENDUDUK_DASHBOARD_ID = int(os.environ.get("MB_PENDUDUK_DASHBOARD_ID", "2"))

DESA_API = os.environ.get("DESA_API", "http://localhost:8081/api/v1")
DESA_USER = os.environ.get("DESA_USER", "admin@desa.local")
DESA_PASS = os.environ.get("DESA_PASS", "admin123")

HERE = os.path.dirname(os.path.abspath(__file__))
DATASETS = os.path.join(HERE, "..", "datasets")

DRY_RUN = False
SESSION = None

# Cards on the Penduduk dashboard that we create and must preserve.
KEEP_CARD_NAMES = {"Piramida Penduduk 2021", "Rasio Ketergantungan Penduduk"}

# Dashboards we own (Penduduk id 2 excluded — user's dashboard).
DASHBOARD_NAMES = [
    "Pendidikan", "Pekerjaan", "Kesehatan", "Pertanian",
    "Keagamaan", "Pemerintahan", "Ekonomi & Infrastruktur",
    "Keluarga Berencana",
]

# desa API: dashboard name -> infographic category name
DESA_CATEGORY = {
    "Pendidikan": "Pendidikan",
    "Pekerjaan": "Statistik",
    "Kesehatan": "Kesehatan",
    "Keluarga Berencana": "Kesehatan",
    "Pertanian": "Statistik",
    "Keagamaan": "Statistik",
    "Pemerintahan": "Statistik",
    "Ekonomi & Infrastruktur": "Keuangan",
}


# ------------------------------------------------------------------- MB HTTP
def api(method, path, data=None, raw=None, content_type=None):
    headers = {}
    if SESSION:
        headers["X-Metabase-Session"] = SESSION
    if raw is not None:
        headers["Content-Type"] = content_type or "application/octet-stream"
        body = raw
    elif data is not None:
        headers["Content-Type"] = "application/json"
        body = json.dumps(data).encode()
    else:
        body = None
    req = urllib.request.Request(MB_URL + path, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=90) as resp:
            b = resp.read()
            return json.loads(b) if b else {}
    except urllib.error.HTTPError as e:
        raise RuntimeError(f"{method} {path} -> {e.code}: {e.read().decode()[:400]}")


def login():
    global SESSION
    SESSION = api("POST", "/api/session", {"username": MB_USER, "password": MB_PASSWORD})["id"]
    print(f"logged in ({MB_USER})")


def upload_csv(path):
    boundary = "----mb" + uuid.uuid4().hex
    fn = os.path.basename(path)
    with open(path, "rb") as f:
        content = f.read()
    ct = mimetypes.guess_type(fn)[0] or "text/csv"
    parts = [f'--{boundary}\r\nContent-Disposition: form-data; name="collection_id"\r\n\r\n{MB_COLLECTION_ID}\r\n'.encode()]
    parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{fn}"\r\nContent-Type: {ct}\r\n\r\n'.encode() + content + b"\r\n")
    parts.append(f"--{boundary}--\r\n".encode())
    api("POST", "/api/upload/csv", raw=b"".join(parts),
        content_type=f"multipart/form-data; boundary={boundary}")


def find_table(prefix):
    meta = api("GET", f"/api/database/{MB_DB_ID}/metadata")
    matches = [t["name"] for t in meta.get("tables", []) if t["name"].startswith(prefix)]
    if not matches:
        raise RuntimeError(f"table {prefix}* not found")
    return sorted(matches)[-1]


def coll_items(model):
    return api("GET", f"/api/collection/{MB_COLLECTION_ID}/items?models={model}").get("data", [])


def find_card(name):
    return next((c for c in coll_items("card") if c.get("name") == name and not c.get("archived")), None)


def find_dashboard(name):
    root = api("GET", "/api/dashboard")
    root = root if isinstance(root, list) else root.get("data", [])
    for d in root:
        if d.get("name") == name and not d.get("archived") and d.get("id") != PENDUDUK_DASHBOARD_ID:
            return d
    return None


# ------------------------------------------------------------------ cleanup
def cleanup():
    print("\n[cleanup]")
    for name in DASHBOARD_NAMES:
        d = find_dashboard(name)
        if not d:
            continue
        if DRY_RUN:
            print(f"  [dry-run] delete dashboard {name} ({d['id']})")
            continue
        try:
            api("DELETE", f"/api/dashboard/{d['id']}")
            print(f"  deleted dashboard {name} ({d['id']})")
        except RuntimeError as e:
            print(f"  dashboard {name}: {e}")

    # delete our cards (all collection-5 cards except user's + kept ones)
    keep = KEEP_CARD_NAMES
    user_ids = set()
    for dc in api("GET", f"/api/dashboard/{PENDUDUK_DASHBOARD_ID}").get("dashcards", []):
        pass
    for c in coll_items("card"):
        name = c.get("name", "")
        if name in keep or name in ("Total Penduduk YoY", "Penduduk By Gender YoY"):
            continue
        if DRY_RUN:
            print(f"  [dry-run] delete card {name} ({c['id']})")
            continue
        try:
            api("DELETE", f"/api/card/{c['id']}")
        except RuntimeError:
            pass
    print("  removed our cards")


def desa_cleanup():
    print("\n[desa API cleanup]")
    token = desa_login()
    headers = {"Authorization": "Bearer " + token}
    req = urllib.request.Request(DESA_API + "/infographic?limit=100", headers=headers)
    with urllib.request.urlopen(req) as r:
        rows = json.load(r)["data"]["infographic"]
    for row in rows:
        if row["section_name"] in DASHBOARD_NAMES:
            if DRY_RUN:
                print(f"  [dry-run] delete infographic {row['section_name']}")
                continue
            d = urllib.request.Request(f"{DESA_API}/infographic/{row['id']}", headers=headers, method="DELETE")
            try:
                with urllib.request.urlopen(d):
                    print(f"  deleted infographic {row['section_name']}")
            except urllib.error.HTTPError as e:
                print(f"  delete {row['name']} failed: {e.code}")


# ------------------------------------------------------------------ card specs
def summary_spec(name, display, where, desc, ylab=None, extra=None):
    sql = f"SELECT indikator, nilai FROM {{derived}} WHERE {where} ORDER BY nilai DESC"
    viz = {}
    if display == "pie":
        viz = {"pie.dimension": "indikator", "pie.metric": "nilai"}
    elif display in ("bar", "line"):
        viz = {"graph.dimensions": ["indikator"], "graph.metrics": ["nilai"]}
        if ylab:
            viz["graph.y_axis.title_text"] = ylab
    elif display == "scalar":
        viz = {"scalar.field": "nilai"}
    elif display == "table":
        viz = {"table.columns": [{"name": "indikator", "enabled": True},
                                 {"name": "nilai", "enabled": True}]}
    if extra:
        viz.update(extra)
    return (name, display, sql, desc, viz)


def hist_spec(name, where, desc, split=False, ylab="Nilai"):
    sql = f"SELECT tahun, indikator, nilai FROM {{hist}} WHERE {where} ORDER BY tahun"
    viz = {
        "graph.dimensions": ["tahun", "indikator"],
        "graph.metrics": ["nilai"],
        "graph.x_axis.title_text": "Tahun Publikasi",
        "graph.y_axis.title_text": ylab,
        "graph.show_values": True,
    }
    if split:
        viz["graph.split_panels"] = True
    return (name, "line", sql, desc, viz)


def build_cards():
    D = "dashboard = '{}'"
    c = {}

    # ---- summary (latest derived) ----
    c["ringkasan_pendidikan"] = summary_spec(
        "Ringkasan: Tingkat Pendidikan (%)", "bar",
        D.format("Pendidikan") + " AND indikator IN ('Belum sekolah','Tamat SD','Tamat SLTP','Tamat SLTA','Akademi','Universitas')",
        "Persentase penduduk menurut jenjang pendidikan tertinggi (2021).", "Persen")
    c["rasio_guru"] = summary_spec(
        "Ringkasan: Rasio Siswa per Guru", "bar",
        D.format("Pendidikan") + " AND indikator LIKE 'Siswa per Guru%'",
        "Rata-rata jumlah siswa per guru, SD vs SMP (2021).", "Siswa")
    c["ringkasan_pekerjaan"] = summary_spec(
        "Ringkasan: Komposisi Pekerjaan (%)", "pie", D.format("Pekerjaan"),
        "Persentase penduduk menurut mata pencaharian (2021).")
    c["ringkasan_kesehatan"] = summary_spec(
        "Ringkasan: Fasilitas & Tenaga per 1.000 Penduduk", "bar",
        D.format("Kesehatan") + " AND indikator LIKE '% per 1.000%'",
        "Ketersediaan tenaga & fasilitas kesehatan per 1.000 penduduk (2021).", "Per 1.000")
    c["kb_aktif"] = summary_spec(
        "Ringkasan: Cakupan KB Aktif (%)", "scalar",
        D.format("Kesehatan") + " AND indikator = 'Cakupan KB Aktif'",
        "Persentase PUS akseptor KB aktif (2021).")
    c["disabilitas_pct"] = summary_spec(
        "Ringkasan: Penyandang Disabilitas (%)", "scalar",
        D.format("Kesehatan") + " AND indikator = 'Penyandang Disabilitas'",
        "Persentase penduduk penyandang disabilitas (2021).")
    c["produktivitas_padi"] = summary_spec(
        "Ringkasan: Produktivitas Padi (ton/Ha)", "scalar",
        D.format("Pertanian") + " AND indikator = 'Produktivitas Padi'",
        "Produktivitas padi (2021).")
    c["ringkasan_lahan"] = summary_spec(
        "Ringkasan: Komposisi Lahan (%)", "pie",
        D.format("Pertanian") + " AND satuan = 'persen'",
        "Persentase penggunaan lahan desa (2021).")
    c["ringkasan_agama"] = summary_spec(
        "Ringkasan: Pemeluk Agama (%)", "pie",
        D.format("Keagamaan") + " AND satuan = 'persen'",
        "Persentase penduduk menurut agama (2021).")
    c["ibadah_rasio"] = summary_spec(
        "Ringkasan: Sarana Ibadah per 1.000 Penduduk", "scalar",
        D.format("Keagamaan") + " AND indikator LIKE 'Sarana%'",
        "Sarana ibadah per 1.000 penduduk (2021).")
    c["ringkasan_pemerintahan"] = summary_spec(
        "Ringkasan: Kelembagaan & Wilayah Desa", "table", D.format("Pemerintahan"),
        "Data kelembagaan & wilayah Desa Palasari (2021).")
    c["pbb_pct"] = summary_spec(
        "Ringkasan: Realisasi PBB (%)", "scalar",
        D.format("Ekonomi & Infrastruktur") + " AND indikator = 'Realisasi PBB'",
        "Persentase realisasi PBB terhadap target (2021).")
    c["akses_pct"] = summary_spec(
        "Ringkasan: Akses Listrik & Air Bersih (%)", "bar",
        D.format("Ekonomi & Infrastruktur") + " AND indikator LIKE 'Akses%'",
        "Persentase rumah tangga dengan akses listrik & air bersih (2021).", "Persen")
    # ---- historical (line) ----
    c["hist_pemerintahan_kk"] = hist_spec(
        "Histori: Jumlah KK", "dashboard = 'Pemerintahan' AND indikator = 'Jumlah KK'",
        "Perkembangan jumlah keluarga/KK Desa Palasari (2021-2024).", ylab="Keluarga")
    c["hist_rt_rw"] = hist_spec(
        "Histori: RT & RW", "dashboard = 'Pemerintahan' AND indikator IN ('Rukun Tetangga (RT)','Rukun Warga (RW)')",
        "Jumlah RT dan RW per tahun. Catatan: nilai tetap (tidak berubah).", split=True, ylab="Unit")
    c["hist_ibadah"] = hist_spec(
        "Histori: Sarana Ibadah", "dashboard = 'Keagamaan' AND indikator IN ('Masjid','Mushola / Langgar')",
        "Perkembangan jumlah masjid & mushola (2021-2024).", split=True, ylab="Unit")
    c["hist_agama"] = hist_spec(
        "Histori: Pemeluk Agama & Nikah", "dashboard = 'Keagamaan' AND indikator IN ('Penduduk Islam','Penduduk Kristen','Penduduk Budha','Nikah')",
        "Perkembangan pemeluk agama & jumlah nikah (2021-2022).", split=True)
    c["hist_disabilitas"] = hist_spec(
        "Histori: Penyandang Disabilitas", "dashboard = 'Kesehatan' AND indikator LIKE 'Tuna%'",
        "Jumlah penyandang disabilitas per jenis (2021-2022).", split=True, ylab="Jiwa")
    c["hist_kb"] = hist_spec(
        "Histori: Akseptor KB per Metode", "dashboard = 'Keluarga Berencana'",
        "PERHATIAN: lonjakan 2021->2022 diduga karena perubahan definisi (akseptor baru vs peserta aktif) — JANGAN dibaca sebagai tren.", split=True, ylab="Akseptor")
    c["hist_lahan"] = hist_spec(
        "Histori: Penggunaan Lahan", "dashboard = 'Pertanian' AND indikator LIKE 'Lahan%'",
        "Luas penggunaan lahan desa per tahun (2021-2022).", split=True, ylab="Ha")
    c["hist_kelompok"] = hist_spec(
        "Histori: Kelompok Tani", "dashboard = 'Pertanian' AND indikator LIKE 'Kelompok tani%'",
        "Jumlah kelompok tani padi & peternakan (2021-2022). Catatan: padi 8->1 kemungkinan revisi definisi.", split=True, ylab="Kelompok")
    c["hist_pln"] = hist_spec(
        "Histori: Keluarga Pelanggan PLN", "dashboard = 'Ekonomi & Infrastruktur' AND indikator = 'Keluarga pelanggan PLN'",
        "Perkembangan keluarga pelanggan PLN (2021-2024).", ylab="Keluarga")
    c["hist_pasar"] = hist_spec(
        "Histori: Sarana Perdagangan", "dashboard = 'Ekonomi & Infrastruktur' AND indikator IN ('Pasar','Minimarket','Rumah makan / kedai')",
        "Perkembangan sarana perdagangan (2021-2024).", split=True, ylab="Unit")
    c["hist_telekom"] = hist_spec(
        "Histori: Transportasi & Telekomunikasi", "dashboard = 'Ekonomi & Infrastruktur' AND indikator IN ('Menara BTS','Operator seluler')",
        "Menara BTS & operator seluler (2021-2024).", split=True, ylab="Unit")
    c["hist_pbb"] = hist_spec(
        "Histori: PBB Target vs Realisasi", "dashboard = 'Ekonomi & Infrastruktur' AND indikator LIKE 'PBB%'",
        "Target dan realisasi PBB (2021-2022).", split=True, ylab="Rupiah")
    return c


DASHBOARDS = {
    "Pendidikan": ["ringkasan_pendidikan", "rasio_guru"],
    "Pekerjaan": ["ringkasan_pekerjaan"],
    "Kesehatan": ["ringkasan_kesehatan", "disabilitas_pct", "hist_disabilitas"],
    "Keluarga Berencana": ["kb_aktif", "hist_kb"],
    "Pertanian": ["produktivitas_padi", "ringkasan_lahan", "hist_lahan", "hist_kelompok"],
    "Keagamaan": ["ringkasan_agama", "ibadah_rasio", "hist_ibadah", "hist_agama"],
    "Pemerintahan": ["ringkasan_pemerintahan", "hist_pemerintahan_kk", "hist_rt_rw"],
    "Ekonomi & Infrastruktur": ["pbb_pct", "akses_pct", "hist_pln", "hist_pasar", "hist_telekom", "hist_pbb"],
}


# ------------------------------------------------------------------- actions
def make_card(name, display, sql, desc, viz, tables):
    ex = find_card(name)
    if ex:
        print(f"    card exists: {name} ({ex['id']})")
        return ex["id"]
    if DRY_RUN:
        print(f"    [dry-run] card: {name} [{display}]")
        return None
    q = sql
    for token, table in tables.items():
        q = q.replace("{" + token + "}", table)
    res = api("POST", "/api/card", {
        "name": name, "description": desc, "collection_id": MB_COLLECTION_ID,
        "dataset_query": {"database": MB_DB_ID, "type": "native",
                          "native": {"query": q, "template-tags": {}}},
        "display": display, "visualization_settings": viz,
    })
    print(f"    card: {name} ({res['id']})")
    return res["id"]


def layout(ids):
    out = []
    for i, cid in enumerate(ids):
        out.append({"id": -(i + 1), "card_id": cid,
                    "row": (i // 2) * 7, "col": (i % 2) * 12,
                    "size_x": 12, "size_y": 7})
    return out


def make_dashboard(name, ids):
    ex = find_dashboard(name)
    if ex:
        print(f"  dashboard exists: {name} ({ex['id']})")
        return ex["id"]
    if DRY_RUN:
        print(f"  [dry-run] dashboard: {name} ({len(ids)} cards)")
        return None
    d = api("POST", "/api/dashboard", {"name": name, "collection_id": MB_COLLECTION_ID})
    did = d["id"]
    api("PUT", f"/api/dashboard/{did}", {"enable_embedding": True, "embedding_params": {}})
    api("PUT", f"/api/dashboard/{did}/cards", {"cards": layout(ids)})
    print(f"  dashboard: {name} ({did}) + embedding enabled")
    return did


# ----------------------------------------------------------------- desa API
def desa_login():
    req = urllib.request.Request(DESA_API + "/auth/login",
        data=json.dumps({"email": DESA_USER, "password": DESA_PASS}).encode(),
        headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req) as r:
        return json.load(r)["data"]["access_token"]


def register_infographics(name_to_id):
    print("\n[desa API registration]")
    token = desa_login()
    headers = {"Authorization": "Bearer " + token}
    req = urllib.request.Request(DESA_API + "/infographic?limit=100", headers=headers)
    with urllib.request.urlopen(req) as r:
        existing = {row["component_id"] for row in json.load(r)["data"]["infographic"]}
    req = urllib.request.Request(DESA_API + "/public/infographic/categories?limit=100")
    with urllib.request.urlopen(req) as r:
        cats = {c["name"]: c["id"] for c in json.load(r)["data"]["categories"]}

    for name, did in name_to_id.items():
        if did is None or did in existing:
            continue
        slug = "/infographic/" + name.lower().replace("&", "dan").replace(" ", "-")
        payload = {"component_id": did, "component_type": "dashboard",
                   "section_name": name, "section_endpoint": slug,
                   "category": cats.get(DESA_CATEGORY.get(name, "Statistik")), "state": True}
        if DRY_RUN:
            print(f"  [dry-run] register {name} -> {did}")
            continue
        r = urllib.request.Request(DESA_API + "/infographic", data=json.dumps(payload).encode(),
            headers={**headers, "Content-Type": "application/json"}, method="POST")
        try:
            with urllib.request.urlopen(r) as resp:
                print(f"  registered: {name} (component {did}) -> {json.load(resp)['data']['id']}")
        except urllib.error.HTTPError as e:
            print(f"  register {name} failed: {e.code} {e.read().decode()[:200]}")


def main():
    global DRY_RUN
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--skip-cleanup", action="store_true")
    ap.add_argument("--skip-register", action="store_true")
    args = ap.parse_args()
    DRY_RUN = args.dry_run

    login()
    if not args.skip_cleanup:
        cleanup()
        try:
            desa_cleanup()
        except Exception as e:
            print(f"  desa cleanup skipped: {e}")

    if DRY_RUN:
        tables = {"derived": "indikator_derived_<ts>", "hist": "indikator_historis_<ts>"}
    else:
        print("\n[upload]")
        for csv in ("indikator_derived.csv", "indikator_historis.csv"):
            upload_csv(os.path.join(DATASETS, csv))
        tables = {
            "derived": find_table("indikator_derived"),
            "hist": find_table("indikator_historis"),
        }
        print(f"  tables: {tables}")

    cards = build_cards()
    print("\n[cards + dashboards]")
    name_to_id = {}
    for dash_name, keys in DASHBOARDS.items():
        print(f"  [{dash_name}]")
        ids = [make_card(*cards[k], tables=tables) for k in keys]
        did = make_dashboard(dash_name, [i for i in ids if i])
        if did:
            name_to_id[dash_name] = did

    if not args.skip_register:
        register_infographics(name_to_id)
    print("\ndone.")


if __name__ == "__main__":
    main()