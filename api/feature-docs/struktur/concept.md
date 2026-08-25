# Struktur — Concept

Organization chart member management. Profile image is gallery-backed.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Lets admins maintain the village organization chart (name, position, contact info, profile image, freeform description). Public endpoints expose the chart to citizens for transparency. Profile images are served via the unified gallery signed-URL pipeline — there are no struktur-scoped media routes.

## Business Rules

- **`Name` required**, ≤ 255 chars.
- **`ProfileImageMediaID` required** (non-nil after `domain.Validate()`). Legacy `profile_image_url` column dropped in 00045 — service must not accept raw URLs.
- **Upload mutex**: `ImageFile` (raw multipart) and `ProfileImageMediaID` (existing gallery UUID) are mutually exclusive — service rejects when both supplied.
- **Public + admin share the same response shape**: `profile_image_url` (deprecated, `omitempty`) plus canonical `profile_media: {media_id, url, thumbnail_url}` object referencing `/api/v1/media/{id}/...?jwt=...`.
- **Delete tolerates missing media**: gallery row is GC'd separately — `struktur_organisasi.profile_image_media_id` becomes NULL via `ON DELETE SET NULL`, row itself deleted by service.
- **Description persistence**: previously a `db:"-"` tag silently dropped at scan. Fixed — column exists in DB (00037), struct uses `db:"description"`, INSERT/UPDATE/SELECT all include it.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/struktur/struktur.go` | `StrukturHandler` struct + methods (package `struktur`) |
| `interface/http/handler/struktur/route.go` | `RegisterRoutes` for `/struktur/*` + `/public/struktur/*` |
| `domain/struktur/struktur.go` | `Struktur` entity + `Validate()` |
| `usecase/struktur/service.go` | CRUD business logic |
| `interface/postgres/struktur.go` | sqlx implementation — **contains broken Update SQL** (see Known Issues) |
| `db/migrations/00011`, `00037`, `00043`, `00045` | table + `description` + media FK + drop legacy URL |

### Service Dependencies

`Service{repo, clock}`. Lightweight — no email, no external integrations, no rate limiter.

### Repository Interface Methods

`usecase/struktur/repository.go`:

- `Create(ctx, *Struktur) error`
- `Update(ctx, *Struktur) error` — **broken SQL at `interface/postgres/struktur.go:136-167`**
- `Delete(ctx, id string) error`
- `GetByID(ctx, id string) (*Struktur, error)`
- `List(ctx, ListStrukturInput) ([]Struktur, int64, error)`

## Glossary

- **Profile media** — the gallery-backed profile image. Two fields on response: deprecated `profile_image_url` (omitempty) and canonical `profile_media: {media_id, url, thumbnail_url}`. Citizens always resolve via the gallery signed-URL (`scope=public`).
- **Upload mutex** — either raw file (service uploads to gallery, gets new UUID) or pre-existing media UUID. Both → 400.
- **Description fix history** — `bugs.md#3`: previously `db:"-"` tag silently dropped Description at scan time. Now `db:"description"`, included in all INSERT/UPDATE/SELECT.

## Related

- `feature-docs/ppid/concept.md` — sibling feature sharing the gallery signed-URL pipeline (different scope, same mint)
- `usecase/gallery/signed_url.go` — signed-URL minting for `profile_media.url`
- `interface/file/local_handler.go` — image storage adapter (legacy, still wired)
- `pkg/handlerutil.ValidateStruct` — struct validation
- `feature-docs/struktur/database.md`
- `feature-docs/struktur/endpoint.md`
