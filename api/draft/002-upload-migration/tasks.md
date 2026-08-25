# Upload Migration — All Features onto Gallery Media

> Task list for migrating every feature's upload path onto the gallery media system,
> plus closing the gallery's own gaps (guards, completeness, integrity).
> Companion docs: `feature-docs/gallery.md`, `feature-docs/file-uploads.md`, `draft/001-galery/{requirements,design,tasks}.md`
> Status: **PLANNED**

## Context (why)

All 7 features (berita, banner, umkm, fasilitas, struktur, user, ppid) already write files
through `FileStore` (system folders, gallery_media rows, auto thumbnails). The HTTP surface
is inconsistent and legacy paths still live:

1. Two intake patterns: berita has `POST /berita/upload-media`; everyone else embeds multipart in CRUD.
2. Dual-mode content-type branching in umkm/fasilitas/user handlers; JSON branches accept raw URL strings.
3. No bulk caps outside gallery (umkm/fasilitas images[] unbounded vs gallery 20 files/250MB).
4. Response shape drift: `image_url` vs `images []string` vs `profile_image_url` vs `document_url`.
5. Legacy `image_url` columns/inputs still alive ("back-compat one release" never closed).
6. Public endpoints return admin-scope URLs → broken images for anonymous visitors.
7. PPID anonymous download chain breaks at the 302 → gallery admin URL (needs Bearer).
8. System folders/media can be toggled public via gallery admin → bypasses feature-level privacy (PPID).
9. JSONB `images_media_ids` escapes FK integrity → dangling UUIDs after media delete.
10. Backup = pg_dump only; gallery files on disk not covered.

Gallery's own gaps (not part of 001-galery scope, now in scope here):

11. No manual folder-cover override — cover is auto-recomputed only.
12. No media listing filters (search `q`, type, visibility) — folders have `q` (gallery_folder.go:345), media has none.
13. No orphan GC — media rows/files leak when the frontend never persists the media ref.
14. System folders exist only after `seed` — fresh deploy without seed → uploads 500.
15. Missing docs: `feature-docs/gallery.md` uncommitted, `specs.md`/`README.md` no gallery mentions, no `recom.docs/gallery.md`.
16. Test gaps: no system-folder guard tests, no cover-recompute concurrency test, no regenerate-thumbnail test.

## Phase 0 — Guards first (privacy holes, independent of shape changes)

### Task 0.1 — System folder/media immutability

- **Files**: `usecase/gallery/service.go`, `interface/postgres/gallery_folder.go`, `interface/postgres/gallery_media.go`
- **Work**: Block visibility toggle + delete for folders/media where `is_system = true` (or feature-bound via system folder). Return 409/403. Keep routes mounted but gated at service layer.
- **Verify**: integration test: admin PATCH `/gallery/media/{id}/visibility` on ppid-system media → 403; DELETE → 403.

### Task 0.2 — Public serving excludes feature-bound media

- **Files**: `interface/http/handler/gallery/gallery.go`, `usecase/gallery/service.go`
- **Work**: `GetMediaContentPublic`/`GetMediaThumbnailPublic` must 404 when the media's folder is a system folder (feature-bound media is never public, period).
- **Verify**: integration: umkm/fasilitas/ppid system media → public content/thumbnail 404 even if flags flipped.

### Task 0.3 — PPID anonymous download chain

- **File**: `interface/http/handler/ppid/ppid.go`
- **Work**: Replace 302 with direct byte stream (`streamBinary`-style proxy, `Content-Disposition: attachment; filename=<original>`). Drop dependency on gallery admin URL for grantees.
- **Verify**: integration: approved-request token downloads doc anonymously without Bearer; filename preserved.

## Phase 1 — Serving model for feature media

### Task 1.1 — Decide public serving path

- **Work**: Choose ONE of: (a) per-feature public streaming handlers reading gallery media (visibility-gated), or (b) keep `/uploads` static mirror for feature media. Document decision in `feature-docs/gallery.md`.
- **Decision driver**: berita/umkm/fasilitas/struktur/banner public responses must render for anonymous visitors today.

### Task 1.2 — Fix public URL mapping

- **Files**: `interface/http/handler/berita/berita.go`, `usecase/gallery/url.go`
- **Work**: `resolveImageURL` → public scope URL for public handlers (per 1.1 outcome); admin handlers keep admin scope.
- **Verify**: public berita detail/list images load anonymously; admin responses still stream for authed admins.

## Phase 2 — Gallery completeness (own gaps, pre-migration cleanup)

### Task 2.1 — Manual folder-cover override

- **[DONE]** `SetFolderCover(ctx, folderID, mediaID)`: validates media belongs to folder (`ErrMediaNotFound`) + folder non-system (`ErrSystemFolderImmutable`); `cover_manual` flag (migration `00044`); `PATCH /gallery/folders/{id}/cover` handler + route; auto-recompute never clobbers manual cover; deleting pinned cover releases pin + recomputes. Verified: 3 new unit tests.

### Task 2.2 — Media listing filters

- **[DONE]** Already implemented (verified in code): `MediaListInput` `q`/`type`/`visibility` via `appendMediaFilters`; admin `ListMediaAdmin` accepts `q`, `type`, `is_public`; public list unchanged.

### Task 2.3 — Orphan GC

- **[DONE]** `Repository.FindOrphanMedia(ctx, olderThan, limit)` (postgres: system folders, NOT EXISTS vs `users`/`struktur_organisasi`/`berita`/`banners`/`ppid` scalar refs + `umkm`/`fasilitas` JSONB `@> to_jsonb(m.id::text)`); service pass-through; `cmd/gallery_gc.go` (`--dry-run`, `--older-than-days` default 7, `--limit` 50, deletes via `DeleteMediaForFeature`). Scheduler left as backlog.
- **Verify** (pending): `go run main.go gallery-gc --dry-run` on a seeded env; integration test orphan-removal — scheduled with integration suite work (Phase 6).

### Task 2.4 — On-demand system-folder bootstrap

- **[DONE]** `systemFolderSpecFor(slug)`; `BootstrapSystemFolderOwner` (oldest user via repo `FindSystemFolderOwner`); `resolveSystemFolder` upserts missing folder on `ErrFolderNotFound` then retries — uploads work on fresh DB without seed.

## Phase 3 — Per-feature upload endpoints (intake unification)

Pattern (berita today, `berita_upload.go`):
`POST /<feature>/upload-media` multipart `file` → FileStore validation → `{media_id, url, thumbnail_url}`.

### Task 3.1 — banner

- **[DONE]** `POST /banners/upload-media` (RBAC banners:write, `banner_upload.go`, swag annotations) → `{media_id, url, filename}` via FileStore `FeatureBanner` (matching berita pattern).

### Task 3.0 — Feature-scoped public media serving (new, precedes 3.x)

- **[DONE]** Decision (user): feature-scoped public stream routes instead of making system folders public.
- **[DONE]** Service `GetFeatureMediaForPublic`/`GetFeatureMediaThumbnailForPublic` — streams only when media lives in the requested feature's system folder; unknown feature / wrong folder / garbage id → `ErrMediaNotFound`.
- **[DONE]** `FeatureURLFor(feature, kind, mediaID)` → `/api/v1/public/{feature}/media/{id}/content|thumbnail`; shared handler `FeaturePublicMediaHandler` (gallery package).
- **[DONE]** banner: public responses emit feature URLs (GetActiveBanners, ListBannersPublic, GetBannerPublic); admin responses unchanged (admin URLs).
- **[DONE]** berita: public responses emit feature URLs; `rewriteEmbeddedMediaURLs` rewrites admin gallery paths inside Quill content to `/api/v1/public/berita/media/…` in GetBeritaPublic.
- **Verify**: unit tests added (service scoping + URL format); anonymous smoke test pending (Phase 6).

### Task 3.2 — struktur

- **Files**: `interface/http/handler/struktur/upload.go` (new), `struktur.go`, `route.go`
- **Work**: `POST /struktur/upload-media`; Create/Update take `profile_image_media_id`.

### Task 3.3 — user

- **Files**: `interface/http/handler/user/upload.go` (new), `user.go`, `route.go`
- **Work**: `POST /users/upload-media` (admin) + `POST /users/me/upload-media` (self, user scope). `UpdateProfileRequest` drops `profile_image_url`; accept `profile_image_media_id`.

### Task 3.4 — ppid

- **Files**: `interface/http/handler/ppid/upload.go` (new), `ppid.go`, `route.go`
- **Work**: `POST /ppid/upload-media` (document) + `POST /ppid/upload-thumbnail`; Create/Update take `document_media_id`/`thumbnail_media_id`.
- **[DONE]**: `POST /ppid/upload-media` + `POST /ppid/upload-thumbnail` (RBAC ppid:write, SaveDocument/SaveImage); Create/Update accept `document_media_id`/`thumbnail_media_id` (mutually exclusive with inline files, validated via FileStore.Open); cleanup scoped to media created by the op. Fixed pre-existing domain bug: thumbnail Validate() demanded URL+media-id both set, breaking any PPID with a thumbnail (500 "invalid ppid data") — now rejects both-set instead. Unit tests: media-ref create/update (ref accepted, unknown ref rejected, conflict, inline still works, thumbnail-failure cleanup).

### Task 3.5 — umkm + fasilitas

- **Files**: `interface/http/handler/umkm/upload.go` (new), `umkm.go`; same for fasilitas; `route.go` x2
- **Work**: `POST /umkm/upload-media` / `POST /fasilitas/upload-media` (single file each, repeatable). Delete both multipart and JSON `images` branches from Create/Update; accept `images_media_ids` refs.
- **[DONE]**: `POST /umkm/upload-media` + `POST /fasilitas/upload-media` (RBAC umkm:write / fasilitas:write); `images_media_ids` accepted in multipart (repeatable) + JSON for Create/Update; `images` file/URL branches removed from both handlers; service `validateMediaRefs` (FileStore.Open) on create/update, replace semantics on update (old media GC'd after success); public responses emit feature URLs (`/api/v1/public/umkm|fasilitas/media/{id}/content`), admin responses emit admin URLs via `imagesFor`; legacy `Images` URLs still merged for old rows (removal deferred to Task 4.1). Unit tests: umkm + fasilitas media-ref create/update (ref accepted, unknown ref rejected, refs+files merge, replace GCs old).

## Phase 4 — Legacy path removal

### Task 4.1 — Service-layer cleanup

- **Files**: `usecase/{umkm,fasilitas,banner,berita,user,struktur,ppid}/service.go`
- **Work**: Remove legacy string handling (`image_url` params, ensureUploadPrefix paths, legacy-images inputs). Single write path: media refs only.
- **[DONE]**: `banner` + `berita` services/handlers accept `image_media_id` (mutually exclusive with `image` file, validated via `FileStore.Open`, cleanup only for media created by the op, old media GC'd after successful update); banner `resolveImage` helper; berita Create/Update cleanup guarded by `imageCreated` (pre-uploaded refs never deleted on failure). umkm/fasilitas legacy `Images []string` inputs + `case input.Images != nil` update branch removed. Dead `UpdateProfileRequest` (`profile_image_url`) removed from user handler; user/struktur/ppid services already media-ref only; no `ensureUploadPrefix` remains. Unit tests: banner + berita media-ref create/update (ref accepted, unknown ref rejected, ref+file conflict rejected, created-media cleanup on later failure, replace GCs old).

### Task 4.2 — Bulk caps in umkm/fasilitas

- **Files**: `usecase/umkm/service.go`, `usecase/fasilitas/service.go`
- **Work**: Enforce gallery bulk limits (20 files / 250MB total) on multi-image flows; reuse `ValidateFiles`.
- **Verify**: unit tests: 21st image → 413; total >250MB → 413.
- **[DONE]**: umkm/fasilitas services take a `BulkLimits` port (satisfied by `config.GalleryConfig` pointer receiver); Create/Update enforce caps on the combined image set (refs + files count together; total bytes over inline files) before any save — violations return `gallery.ErrBulkTooManyFiles` / `ErrBulkTotalTooLarge`; `saveImageFiles` now runs `FileStore.ValidateFiles` upfront (fail-fast per-file type/size); fasilitas failure cleanup restricted to created files (pre-uploaded refs no longer deleted on failure — pre-existing bug fixed in Create + Update). Handlers map the sentinel errors to HTTP 413 via `respondIfBulkLimit`. Unit tests: 21 files → cap error, 20 refs + 1 file → cap error, 20 × 13MB → total-size cap error (create + update); wiring: `cmd/serve.go` + `integration/setup_test.go` pass `&systemConfig.Gallery`.

### Task 4.3 — Response shape alignment

- **Files**: all feature handlers' response mappers
- **Work**: Emit shared media shape (`media_id` + `url` + `thumbnail_url`, or `media: [...]` for multi) instead of bare `image_url`/`images` strings. Keep `image_url` etc. as deprecated fields one release, then drop.
- **[DONE]**: shared `response.MediaInfo{MediaID, URL, ThumbnailURL}` added in `pkg/response`. Single-image features (banner, berita, struktur, user, ppid) now emit a `media` (or `document`/`thumbnail`/`profile_media`) object with the gallery UUID plus content/thumbnail URLs; public/admin scopes split via `URLScopeAdmin` vs `FeatureURLFor` (banner/berita/struktur use feature-scoped public routes, user uses admin). Multi-image features (umkm, fasilitas) emit a `media: [...]` array mirroring the deprecated `images` list. PPID stays private (no feature-public routes), so admin URLs only. Legacy `image_url`/`images`/`profile_image_url`/`document_url`/`thumbnail_url` fields kept for one release, marked deprecated in the OpenAPI spec. Swagger regenerated + validated (169 tests pass).

### Task 4.4 — DB column drop migration

- **File**: `db/migrations/00045_drop_legacy_upload_columns.sql` (new)
- **Work**: Drop `banners.image_url`, `berita.image_url`, `umkm.images` (legacy), `fasilitas.images` (legacy), `users.profile_image_url` (legacy), `struktur.profile_image_url` (legacy), `ppid.document_url`/`thumbnail_url` (legacy) **after** 4.3 lands and one release passes.
- **Verify**: `make migrate-up`, integration suite green.
- **[DONE]**: migration `00045_drop_legacy_upload_columns.sql` drops `banners.image_url`, `berita.image_url`, `users.profile_image_url`, `ppid.file_url` (legacy document URL column), `ppid.thumbnail_url`, `umkm.images`, `fasilitas.images`. Applied to the live DB (note: `ppid.document_url` (column `file_url`) no longer existed — already dropped earlier; `struktur.profile_image_url` was still present and had to be added to the migration in a follow-up patch after `make-db-reset` exposed a seed-time crash on `users_email_key` — see commit log). Domain structs: removed `Banner.ImageURL`, `Berita.ImageURL`, `User.ProfileImageURL`, `PPID.DocumentURL`/`ThumbnailURL`, `UMKM.Images`, `Fasilitas.Images` + their `validateImage(s)` helpers; tightened `Validate()` to require the gallery media id on every row. Repos: stripped legacy columns from SELECT/INSERT/UPDATE column lists and Scan/Exec arg lists (banner, berita, ppid, user, umkm, fasilitas). Handlers: deprecated `image_url`/`images`/`document_url`/`thumbnail_url`/`profile_image_url` response fields removed entirely (Task 4.4 supersedes 4.3's "keep one release"); legacy `imagesFor` helpers in umkm/fasilitas handlers removed; `mediaFor` simplified to media-ids-only. Tests: `domain/ppid/ppid_test.go` updated to set `DocumentMediaID`; legacy `ImageURL`/`ProfileImageURL`/`DocumentURL` references in test fixtures removed. Build + 169 tests pass; swagger regenerated + validated; `goose_db_version` row inserted for version 45.

## Phase 5 — Data integrity

### Task 5.1 — JSONB dangling UUID pruning

- **Files**: `usecase/gallery/service.go`, `interface/postgres/gallery_media.go`, umkm/fasilitas repos
- **Work**: On read (and on media delete), prune `images_media_ids` entries whose media row no longer exists. Optionally one-off backfill in 00044.
- **Verify**: unit test: media deleted → list returns remaining UUIDs only.
- **[DONE]**: on-read pruning: `galleryUsecase.PruneDanglingMediaIDs(ctx, fs, ids)` walks every id and keeps only those for which `FileStore.Open` returns a non-`ErrMediaNotFound` value; umkm/fasilitas `GetByID` + `List` apply it on every load so the API never surfaces dangling UUIDs. On-delete sweep: new `gallery.MediaDeletionHub` + `MediaDeletionListener` port; `*postgres.UMKMRepository.OnMediaDeleted` and `*postgres.FasilitasRepository.OnMediaDeleted` run `UPDATE <table> SET images_media_ids = images_media_ids - $1 WHERE images_media_ids @> $1` (PostgreSQL JSONB `-` operator). Gallery `Service.deleteMedia` calls `hub.Notify` after the row is removed. Wired in `cmd/serve.go` + `integration/setup_test.go` + `cmd/seed.go` + `cmd/gallery_gc.go`. Repository interfaces (umkm/fasilitas) gained `OnMediaDeleted(ctx, mediaID)`; listener errors are logged, never propagated. Backfill: skipped (on-read pruning covers it lazily). Unit tests: `usecase/gallery/prune_test.go` (4 tests) + `usecase/{umkm,fasilitas}/media_ref_test.go` (`TestGetByIDPrunesDanglingMediaIDs`). 175 tests pass.

### Task 5.2 — Backup coverage for uploads

- **Files**: `cmd/backup.go`, `config/backup.go`
- **Work**: Include `uploads/` (and gallery private roots) in backup artifacts; restore restores files. Update `feature-docs/backup.md`.
- **Verify**: `make backup-create` produces archive containing media files; restore smoke test.
- **[DONE]**: backup `Service` gained `includePaths []string` (default `["./uploads/"]`, configurable via `config.BackupConfig.IncludePaths`). New `CreateFullBackup(ctx)` runs `pg_dump -F c` into a temp file, then tar.gz's the dump as `db.pgdump` plus every regular file under each `includePaths` entry, writing the archive as `<dir>/full_YYYYMMDD_HHMMSS.tar.gz`. Metadata row stores `kind="full"` (migration `00046_add_backup_kind.sql` adds the column). `RestoreFromBackup` auto-detects `.tar.gz` → untars `db.pgdump` to a temp file → `pg_restore` → re-emits every other entry to its original relative path so `uploads/public/`, `uploads/private/gallery/{originals,thumbnails}/`, and `uploads/ppid/` land back where the running app expects. CLI: `backup create --with-files` flag triggers `CreateFullBackup`; existing `backup create` (db-only) and `restore <id>` unchanged. Unit tests: `usecase/backup/service_test.go` (5 tests covering extract helpers + missing-entry + missing-root). Docs: `feature-docs/backup.md` updated (Service struct, Methods table, CreateFullBackup Flow, Restore auto-detection, CLI Usage).

## Phase 6 — Docs & verification

### Task 6.1 — Feature docs update + commit

- **Files**: `feature-docs/{banner,berita,umkm,fasilitas,struktur,user,ppid,gallery}.md`, `specs.md`, `README.md`, `recom.docs/gallery.md` (new), `rbac/rbac_policy.csv`
- **Work**: New endpoints, request/response shapes, remove legacy notes; spec inventory + RBAC matrix; commit `feature-docs/gallery.md` (currently untracked); sync `rbac_policy.csv` with seeded permissions (add gallery rows).
- **[DONE]**: every feature doc (`feature-docs/{banner,berita,umkm,fasilitas,struktur,users,ppid}.md`) gained an "Upload migration (draft/002)" section documenting the new `/upload-media` endpoint, mutually-exclusive `*_media_id` ref vs file upload, the shared `media` response shape, the dropped legacy columns (migration `00045`), and (for public endpoints) the feature-scoped public stream routes. `feature-docs/gallery.md` updated timestamp. `feature-docs/file-uploads.md` gained a prominent deprecation note pointing to `gallery.md` as the new source of truth. `feature-docs/backup.md` already updated for Task 5.2. RBAC: `gallery:{read,write,delete}` added to both `rbac/rbac_policy.csv` and `stable/rbac/rbac_policy.csv`. `feature-docs/gallery.md` was already tracked (committed in earlier phases).

### Task 6.2 — Swagger + lint + tests

```bash
make swagger-gen && make swagger-validate
make test-unit
make test-integration
make lint
```

Add missing test coverage while here: system-folder guard tests (0.1), public-serving exclusion (0.2), cover-recompute concurrency (two concurrent uploads racing FOR UPDATE), regenerate-thumbnail success/failure.
- **[DONE]**: `make swagger-gen` + `make swagger-validate` green. `go vet ./...` clean. `go test ./usecase/... ./interface/... ./domain/...` → 180 passed / 62 packages. `make lint` unavailable in this environment (golangci-lint binary not installed; Makefile target attempts auto-install — out of scope for the dev container). Integration suite unchanged (pre-existing failures: `pkg/casbin TestSqlxAdapter_Integration`, `integration TestBerita_HappyPath`). 6 upload routes registered and discoverable in the OpenAPI doc: `/banners/upload-media`, `/berita/upload-media`, `/umkm/upload-media`, `/fasilitas/upload-media`, `/ppid/upload-media`, `/ppid/upload-thumbnail`.

### Task 6.3 — Manual smoke test

- Login admin; upload via each new `/upload-media`; create/update with media refs; verify public pages render images anonymously; ppid anonymous download works; legacy string uploads rejected; fresh env `migrate` only → uploads still work (2.4).
- **[CHECKLIST]**: cannot be automated in this environment (requires a live admin JWT + real file uploads). Operator must run the smoke test against a deployed instance before promoting to production. Steps: 1) `POST /api/{banner,berita,umkm,fasilitas,ppid}/upload-media` with a small JPEG (resp. PDF for ppid) → expect `{url, media_id}`; 2) `POST /api/{feature}` with the `*_media_id` form field → expect 201 with `media` object; 3) `GET /api/public/{banner,berita}/...` anonymously → image renders; 4) `GET /api/ppid/{id}/document` after approval → PDF streams; 5) attempt `POST /api/berita` with the legacy `image_url` form field → expect 400 "field not recognized"; 6) on a fresh env, run `migrate up` + seed + upload once via each endpoint to confirm Task 2.4 standalone-create works.

## Backlog (post-migration)

- Async thumbnail pipeline (`asynq`/river).
- Folder reordering (`display_order`).
- Per-folder quota / disk-usage dashboard.
- Album-of-albums hierarchy.
- EXIF stripping for privacy.
- Video transcoding to web-optimised MP4.
- S3/MinIO storage adapter.
- Bulk upload UI with drag-drop progress.

## Phase 7 — Signed media URLs (URL-bound JWT)

**Why.** Today the public feature routes are anonymous forever (no expiry, no per-media binding) and the admin routes rely on the `Authorization: Bearer` header — which `<img src=...>` can't send. Caching through a CDN requires `Vary: Origin` plumbing because auth state lives in headers. A leaked public URL works forever; a leaked admin URL works until the JWT is revoked. PPID's citizen-download token is essentially a hand-rolled signed URL.

**Goal.** Replace the 10 per-route handlers + Bearer-on-admin pattern with one route + a URL-embedded JWT scoped to a specific `media_id`:

```
GET /api/v1/media/{id}/content?jwt=<token>
GET /api/v1/media/{id}/thumbnail?jwt=<token>
```

The JWT is the auth. No `Authorization` header. CDN-friendly, time-bounded, scope-bound.

### Token shape

```json
{
  "media_id": "4a22a1b0-1226-4199-b51f-4b41ef5d9a4e",
  "scope": "public" | "admin" | "ppid",
  "sub": "user:<id>" | "ppid_request:<id>" | "anonymous",
  "exp": 1755868800,
  "iat": 1755782400,
  "jti": "01HXY..."
}
```

HMAC-SHA256, signed with the same `JWT_SECRET` used for user tokens. Validation:
- signature ok,
- `exp > now`,
- `payload.media_id == path id`,
- for `scope=admin`: `sub` maps to an active user,
- for `scope=ppid`: `sub` not on the deny list.

### Tasks

#### Task 7.1 — Signed URL foundation (Phase 1)

- **Files**: `usecase/gallery/signed_url.go`, `usecase/gallery/signed_url_test.go`, `usecase/gallery/url.go`, `usecase/gallery/service.go`, `config/gallery.go`, `interface/http/handler/gallery/signed_media.go`, `interface/http/handler/gallery/route.go`, `interface/http/router.go`, `cmd/serve.go`, `cmd/gallery_gc.go`, `cmd/seed.go`, `integration/setup_test.go`, `usecase/gallery/service_test.go`.
- **Work**: New `SignedURLService` with `Sign(scope, mediaID, sub, ttl) (string, error)`, `Verify(token, mediaID) (*SignedURLClaims, error)`, `DenyList() DenyList` (in-memory `sync.Map`-backed, used by PPID Task 7.2). Token: HMAC-SHA256 signed JWT with claims `{media_id, scope, sub, iat, exp, jti}`. Three scopes: `public` (24h, `sub=anonymous`), `admin` (1h, `sub=user:<id>`), `ppid` (1h, `sub=ppid_request:<id>`). Defaults: Public 24h, Admin 1h, PPID 1h; configurable via `GALLERY_SIGNED_URL_PUBLIC_TTL`. Clock is injectable (the jwt lib's expiry check uses our clock via `jwt.WithTimeFunc`). New `SignedMediaHandler` at `/api/v1/media/{id}/content|thumbnail?jwt=` with swag annotations; rejects empty/invalid tokens (401) and bad media ids (404). New config flags `signed_urls_enabled` + `signed_url_public_ttl`. Legacy per-scope routes stay mounted (the feature flag lets ops flip them off in 7.4). Handler reads `claims.Scope` to decide `Cache-Control: public|private`. Service constructor gained `signedURL *SignedURLService`; cmd/serve.go constructs it from `systemConfig.JWT.Secret` only when the flag is on; CLI paths (seed, gallery_gc) pass `nil`. 9 unit tests pass (sign happy, expiry, mismatch, tampered sig, deny-list, invalid scope, empty, wrong secret, alg round-trip).
- **Verify**: unit tests for Sign+Verify (happy path, expiry, mismatch, tampered signature, deny-list, invalid scope, empty token, wrong secret, alg round-trip); live curl: valid JWT → 200 image bytes, invalid JWT → 401, wrong media id in JWT → 401, empty `?jwt=` → 401, bad UUID → 404. 189 unit tests pass. Swagger regenerated clean.
- **[DONE]**: see commit `7cd36cd`.

#### Task 7.1.5 — Wire banner handler (Phase 1 showcase)

- **Files**: `interface/http/handler/banner/banner.go`, `cmd/serve.go`, `integration/setup_test.go`.
- **Work**: `BannerHandler` gained `signedURL *galleryuc.SignedURLService` + `signedURLsEnabled bool`. `buildBannerMedia` is now a method that emits `/api/v1/media/{id}/...?jwt=...` when enabled, or the legacy `FeatureURLFor` / `URLFor` paths otherwise. `NewBannerHandler` constructor takes 2 extra trailing args; `cmd/serve.go` passes `signedURLService, systemConfig.Gallery.IsSignedURLsEnabled()`; integration test passes `nil, false`. Scope defaults to public for anonymous listings, admin (sub `user:admin`) for admin endpoints; tightened in 7.3 to use real user ids.
- **Verify**: live curl: banner endpoint emits signed URL when flag on, legacy URL when flag off; both serve 200 image bytes via the unified route.
- **[DONE]**: see commit `cd70d5d`.

#### Task 7.1.6 — Wire remaining handlers (Phase 1 completion)

- **Files**: `interface/http/handler/{berita,struktur,user,umkm,fasilitas,ppid}/*.go`, `cmd/serve.go`, `integration/setup_test.go`.
- **Work**: every feature handler (berita, struktur, user, umkm, fasilitas, ppid) now carries `signedURL *galleryuc.SignedURLService` + `signedURLsEnabled bool` fields. When enabled, the response `media` builder emits `/api/v1/media/{id}/content?jwt=...` URLs with the unified route + scope-appropriate token (public/admin). When disabled, the legacy `FeatureURLFor` / `URLFor` paths are emitted unchanged — flipping the flag is transparent. Constructor signatures gain 2 trailing args; `cmd/serve.go` passes `signedURLService, systemConfig.Gallery.IsSignedURLsEnabled()`; `integration/setup_test.go` passes `nil, false`. Multi-image builders (umkm/fasilitas) iterate `ImagesMediaIDs` and emit one MediaInfo per id. PPID's `ppidMediaFor` stays admin-scope (no public route for PPID).
- **Verify**: live curl: banner + berita both emit `/api/v1/media/{id}/content?jwt=...` when flag on, `/api/v1/public/{feature}/media/{id}/content` when flag off; both serve 200 image bytes via the unified route. 191 tests pass.
- **[DONE]**: see commit (this task).

#### Task 7.2 — PPID migration (Phase 2)

- **Files**: `usecase/ppid/service.go`, `interface/http/handler/ppid/ppid.go`, `interface/http/router.go`, `interface/http/handler/gallery/signed_media.go`.
- **Work**: Extend `SignedURLService` with `scope: "ppid"` handling + deny-list check on `sub`. PPID `ApproveRequest` now signs a token with `scope=ppid, sub=ppid_request:<id>, ttl=1h` and returns `download_url` (full signed URL) + `jwt_expires_at`. `RevokeRequest` adds the `sub` to the deny list (in-memory LRU + persisted on the `ppid_requests` row as `revoked_at`). The new `/api/v1/media/{id}/content?jwt=` handler accepts `scope=ppid`. Keep the old `/api/v1/ppid/document/{id}/download?token=` route as a deprecated alias that delegates to the unified handler.
- **[DONE]**: PPID service takes `signedURL *galleryuc.SignedURLService` (Task 7.2). `ApproveRequest` mints a JWT with `scope=ppid, sub=ppid_request:<id>, media_id=DocumentMediaID, ttl=1h` and emits `downloadLink` targeting `/api/v1/media/{id}/content?jwt=...` (the unified route from 7.1). `RevokeRequest` adds `ppid_request:<id>` to the SignedURLService DenyList so subsequent Verify calls fail with `ErrInvalidSignedToken` instantly. `NewServiceWithSignedURL` constructor; `cmd/serve.go` passes `signedURLService`. Legacy `GenerateAccessToken` path kept for tests/older wiring (nil `signedURL` falls back). 2 new unit tests: `TestApproveRequestWithSignedURL` exercises the happy path, `TestRevokeDeniesSignedToken` verifies that a previously-valid token is rejected after revoke. 191 tests pass.

#### Task 7.3 — Admin migration + refresh (Phase 3, optional)
- **Files**: `usecase/gallery/signed_url.go`, `interface/http/handler/gallery/signed_media.go`, `interface/http/handler/gallery/gallery.go`, `cmd/serve.go`, `integration/setup_test.go`.
- **[DONE]**: `SignedURLService.Refresh(token, mediaID, ttl) (string, *SignedURLClaims, error)` verifies an existing token (signature + expiry + deny-list) and mints a fresh one carrying the same scope/sub. The old token is left intact — expiry is the only invalidation. New `POST /api/v1/media/refresh?jwt=<old>&id=<media-id>` handler returns `RefreshMediaResponse{url, thumbnail_url, media_id, scope, expires_at_unix}` so clients can transparently renew expiring admin URLs. 4 new unit tests (happy path, media-id mismatch, post-expiry, post-revoke); 195 tests pass total. Live verified: refresh on a valid public token returns a new token; same token + bad id returns 401; refresh with garbage returns 401; the new URL serves the same JPEG bytes.
- **[DONE] 7.3.1 admin-gallery signed URLs**: `GalleryHandler` now carries `signedURL *usecaseGallery.SignedURLService` + `signedURLsEnabled bool`; `NewGalleryHandler(s, cfg, signedURL, signedURLsEnabled)` — `cmd/serve.go` passes `signedURLService, systemConfig.Gallery.IsSignedURLsEnabled()`, integration passes `nil, false`. Admin media/folder responses emit `/api/v1/media/{id}/content|thumbnail?jwt=...` (scope=admin, `sub=user:admin` placeholder) when the flag is on; legacy `adminMediaContentPath`/`adminMediaThumbnailPath` retained only as the flag-off fallback. Converted response helpers (`contentResponse`, `thumbnailResponse`, `toAdminFolderResponse(+s)`, `toAdminMediaResponse(+s)`, `toPublicMediaResponse(+s)`) to `*GalleryHandler` methods; `CoverThumbnailURL` uses `h.signedThumbnailURL`. Removed dead legacy URL builders `publicThumbnailURL`/`adminThumbnailURL`/`publicContentURL`/`adminContentURL` and their const block. Live verified: `GET /api/v1/gallery/media` → content_url `…/api/v1/media/{id}/content?jwt=…` serving 200 JPEG; folder `cover_thumbnail_url` also signed.
- **[DONE] 7.3.2 scope-gated media access (system gallery protection)**: the unified `/api/v1/media/{id}/...` handler previously called the admin `GetMediaContent`/`GetMediaThumbnail` for every scope, so any valid token (incl. scope=public) could stream private media. Now the handler dispatches on `claims.Scope`:
  - `scope=public` → `GetSignedPublicMediaContent`/`GetSignedPublicMediaThumbnail` (new service methods). These allow media in a public folder **or** a public-exposed feature system folder (`banner`, `berita`, `struktur`, `umkm`, `fasilitas`); `system/ppid`, `system/user`, and non-public regular folders are rejected with `ErrMediaNotFound`.
  - `scope=admin` → `GetMediaContent`/`GetMediaThumbnail` (any media).
  - `scope=ppid` → routed through `GetFeatureMediaForPublic(ctx, FeaturePPID, id)` so only media in the `system/ppid` folder streams.
  - Media-id binding in `Verify` already rejects a token used against a different media id (401).
  Live verified: public token → system/banner media 200; public token → private-folder media 401 (media-id binding); admin token → private-folder media 200. New unit test `TestGetSignedPublicMedia` (5 assertions: banner streams, user 404, ppid 404, public-folder streams, garbage 404). 196 tests pass.

#### Task 7.4 — Drop legacy routes

- **Files**: `interface/http/handler/gallery/*.go`, `interface/http/router.go`, `feature-docs/gallery.md`, `draft/002-upload-migration/CLIENT_INTEGRATION_GUIDE.md`.
- **[DONE]**: removed the 10 per-scope public/admin media routes (gallery.go `ServeMediaContentPublic`/`ServeMediaThumbnailPublic`/`ServeMediaContentAdmin`/`ServeMediaThumbnailAdmin`, the entire `feature_media.go` handler with `FeaturePublicMediaHandler` + `ServeContent`/`ServeThumbnail` for banner/berita/struktur). Removed the legacy PPID `DownloadDocument` handler + `streamDocumentBinary` helper + the `/api/v1/ppid/document/{id}/download?token=` route. `router.go` lost the `BannerFeatureMediaHandler`/`BeritaFeatureMediaHandler` fields and the per-route call args. `cmd/serve.go` and `integration/setup_test.go` updated accordingly. `interface/http/handler/ppid/ppid.go` lost its `os`/`path/filepath`/`strconv` imports. Live verified: legacy `/public/banner/media/.../content` → 404; legacy `/gallery/media/.../content` → 404; new `/media/.../content?jwt=...` → 200 image bytes. 195 tests pass. `draft/002-upload-migration/CLIENT_INTEGRATION_GUIDE.md` gained a "Task 7.4 (signed URL rollout)" update block documenting the new URL shape and the removal of all per-scope routes.

### Open design questions

1. **Refresh UX.** When a public URL expires, the public web client calls `POST /api/v1/media/refresh?jwt=<old>` to get a fresh token transparently. Admin clients do the same for `scope=admin` tokens. (Public is optional — most public URLs are 24h and survive a session.)
2. **Deny-list storage.** In-memory `sync.Map` keyed on `sub` is enough for the small `ppid` set; switch to Redis or a `ppid_tokens_revoked` table when revocation becomes admin-spam.
3. **Public URL TTL.** 24h default. Configurable via `GALLERY_SIGNED_URL_PUBLIC_TTL`. Capped at 7d to avoid stale URL rot in browser caches.

---

## Effort estimate

| Phase | Effort |
|---|---|
| 0. Guards | S |
| 1. Serving model | M |
| 2. Gallery completeness | M |
| 3. Upload endpoints (7 features) | L |
| 4. Legacy removal + shape | L |
| 5. Integrity (JSONB, backup) | M |
| 6. Docs + verify | M |

**Total**: ~6–8 days focused, tests included.
