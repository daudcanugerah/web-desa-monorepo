# UAT Checklist — Webdesa Admin Dashboard

User Acceptance Testing checklist for the admin frontend (`webdesa/admin`).

**File:** `qa/UAT_CHECKLIST.csv` (161 test cases, ~28 KB)
**Generator:** `qa/generate_uat.py` (run with `python3 qa/generate_uat.py` to regenerate).

---

## How to use

1. Open `qa/UAT_CHECKLIST.csv` in Excel / Google Sheets / LibreOffice.
2. Fill the **Status** column as you execute each test: `Pass`, `Fail`, `Blocked`, or `Skipped`.
3. Optional: add a Tester column / Date column if your team tracks them.
4. For each failure, capture the URL, network log, console log, and screenshot in the bug tracker.

The **Status** column starts empty for every row — the CSV is a blank template, not a pre-filled result.

---

## Columns

| # | Column | Purpose |
|---|--------|---------|
| 1 | `ID` | Unique test ID (`<MODULE>-NNN`) — stable across regenerations |
| 2 | `Module` | Feature module (Authentication, Users, News, ...) |
| 3 | `Feature` | Sub-area within the module |
| 4 | `Test Scenario` | Short human title |
| 5 | `Pre-conditions` | What must be true before running |
| 6 | `Steps` | Numbered steps to execute |
| 7 | `Expected Result` | What success looks like |
| 8 | `Priority` | `High` / `Medium` / `Low` |
| 9 | `Type` | `Functional` / `Validation` / `UI` / `UX` / `Security` / `Integration` / `Error Handling` / `i18n` / `Routing` / `Smoke` |
| 10 | `Status` | `Pass` / `Fail` / `Blocked` / `Skipped` (filled during testing) |

---

## Coverage summary

161 test cases across **17 modules**:

| Module | Cases |
|--------|-------|
| Authentication | 16 |
| UI / Cross-cutting | 19 |
| News / Berita | 16 |
| PPID | 15 |
| Fasilitas | 14 |
| Users | 13 |
| UMKM | 9 |
| Infographic | 8 |
| Security / RBAC | 8 |
| Banners | 7 |
| Roles | 7 |
| Profile Sections | 6 |
| My Profile | 5 |
| Struktur | 5 |
| Village Profile | 5 |
| Smoke / Integration | 5 |
| Dashboard | 3 |

By priority:

| Priority | Count |
|----------|-------|
| High | 95 |
| Medium | 47 |
| Low | 19 |

By type:

| Type | Count |
|------|-------|
| Functional | 91 |
| Validation | 21 |
| UI | 17 |
| Security | 10 |
| Smoke | 5 |
| UX | 4 |
| Integration | 4 |
| Error Handling | 4 |
| i18n | 3 |
| Routing | 2 |

---

## Recommended UAT execution order

1. **Smoke first** (SMOKE-001 → SMOKE-005) — verifies the entire login → navigate → logout loop works.
2. **High-priority happy paths** — every `High`-priority `Functional` test.
3. **High-priority validations** — every `High`-priority `Validation` test.
4. **Security / RBAC** — operator cannot access admin-only surfaces.
5. **Category manager modal** — exercises the new `CategoryManagerModal` flow on News, UMKM, PPID.
6. **UI / i18n / Theme** — last, since they don't block any feature.

---

## Regenerate

```bash
python3 qa/generate_uat.py
```

The generator script is the source of truth — add or edit rows in `generate_uat.py` and re-run to refresh the CSV. Quoting (`csv.QUOTE_ALL`) preserves commas and newlines in the Steps / Expected Result cells.

---

## Notes

- The checklist reflects the **post-fix** state of the admin frontend: PPID download uses the singular path, facility image delete uses `DELETE /fasilitas/{id}/images/{idx}`, 429 handling is in place, and category CRUD is wired via `CategoryManagerModal`.
- `/banners/active`, `/infographic/sections/names`, `/permissions`, and `*-categories` endpoints for `banners` / `fasilitas` / `infographic` are not exercised here — see `feature-docs/SWAGGER_DIFF.md` for why.
- "Smoke" appears as both a `Module` (`Smoke`) and a `Type` (`Smoke`) — they are independent dimensions.