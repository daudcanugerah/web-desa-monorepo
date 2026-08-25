# OpenAPI Specification Consistency

Audit results comparing `openapi.yaml` against actual implementation in handlers and routes.

**Audit date:** 2026-07-04
**Status:** All CRITICAL issues resolved by migrating from hand-written `openapi.yaml` to code-generated `docs/swagger.yaml` via [swaggo/swag](https://github.com/swaggo/swag). See [Architecture Option A](./architecture.md#option-a--generate-openapi-from-code-annotations) for implementation.

**OpenAPI version:** 2.0 (swag default)
**OpenAPI file:** `/Users/ivosights/Office/3web/webdesa/api/docs/swagger.yaml` (auto-generated, ~115 KB)
**Router:** `/Users/ivosights/Office/3web/webdesa/api/interface/http/router.go` (351 lines)

---

## Summary

| Metric | Count |
|---|---|
| Active paths in router.go | 111 |
| Methods documented in OpenAPI | 97 |
| **Coverage** | **~87%** |
| Missing admin GET endpoints | 8 |
| CRITICAL issues | 6 |
| MAJOR issues | 18 |
| MINOR issues | 14 |
| **Total discrepancies** | **53** |

Only 4 component schemas defined (`Error`, `Pagination`, `PPID`, `PPIDRequest`). **17+ entity response schemas missing**.

---

## Severity Legend

- 🔴 **CRITICAL** — Wrong path, wrong shape, or invalid spec → will cause consumer failures
- 🟠 **MAJOR** — Missing endpoint, missing required field, missing security → incomplete API surface
- 🟡 **MINOR** — Missing optional fields, missing error responses, slight inconsistencies

---

## 🔴 CRITICAL Issues (6)

### 1. Invalid Path String in OpenAPI

**File:** `openapi.yaml:1498`

**Issue:** Path contains literal space and parentheses:
```yaml
/api/v1/public/ppid/categories (legacy)
```

**Impact:** Invalid HTTP path. Many OpenAPI tools will reject or mangle this path. Generated clients will fail.

**Fix:**
```yaml
# Remove this path entirely
# OR move to a deprecated extension:
x-deprecated: true
deprecated: true
```

---

### 2. Wrong Path: Berita Categories POST

**File:** `openapi.yaml:638`, `interface/http/router.go:216`

**Issue:** OpenAPI documents `POST /api/v1/public/berita/categories` but the actual protected route is `POST /api/v1/berita/categories`.

**Fix:** Change OpenAPI path to `/api/v1/berita/categories` and add `security: [{bearerAuth: []}]`.

---

### 3. Wrong Path: UMKM Categories POST/DELETE

**File:** `openapi.yaml:887, 919`, `interface/http/router.go:233-234`

**Issue:** OpenAPI documents `POST/DELETE /api/v1/public/umkm/categories[/{id}]` but actual routes are `POST/DELETE /api/v1/umkm/categories[/{id}]`.

**Fix:** Change paths and add bearerAuth to both.

---

### 4. Wrong Path: PPID Categories POST

**File:** `openapi.yaml:1145`, `interface/http/router.go:271`

**Issue:** Same pattern — documented under `/public/ppid/categories` but actual is `/ppid/categories`.

**Fix:** Correct path + add bearerAuth.

---

### 5. Fasilitas POST/PUT Body Schema Totally Wrong

**File:** `openapi.yaml:982-1009`, `interface/http/handler/fasilitas.go:58-129`

**Issue:** OpenAPI documents:
```yaml
required: [name, description]
properties:
  name: string
  description: string
  location: object    # ← doesn't exist in Go
```

Actual handler requires:
```go
Name      string                 // required
Latitude  float64                // required, -90 to 90
Longitude float64                // required, -180 to 180
Category  *string                // optional UUID
Description *string              // optional
Images    []string               // optional
```

**Fix:**
```yaml
required: [name, latitude, longitude]
properties:
  name: string
  latitude: number
    format: double
    minimum: -90
    maximum: 90
  longitude: number
    format: double
    minimum: -180
    maximum: 180
  category: string
    format: uuid
  description: string
  images:
    type: array
    items: string
    format: uri
```

---

### 6. Integration Tests Use Wrong PPID Download Path

**File:** `integration/ppid_approval_test.go:349, 391, 405, 423, 638`

**Issue:** Tests use `/api/v1/public/ppid/documents/{id}/download` (plural "documents", public prefix). Actual route is `/api/v1/ppid/document/{id}/download` (singular, no public prefix).

**Impact:** Tests would return 404.

**Fix:** Update tests to use `/api/v1/ppid/document/{id}/download`.

---

## 🟠 MAJOR Issues (18)

### 7-14. Eight Admin GET Endpoints Not Documented

| Endpoint | Where in code |
|---|---|
| `GET /api/v1/berita` | `router.go:204` |
| `GET /api/v1/berita/{id}` | `router.go:205` |
| `GET /api/v1/umkm` | `router.go:223` |
| `GET /api/v1/umkm/{id}` | `router.go:224` |
| `GET /api/v1/fasilitas` | `router.go:240` |
| `GET /api/v1/fasilitas/{id}` | `router.go:241` |
| `GET /api/v1/struktur` | `router.go:278` |
| `GET /api/v1/struktur/{id}` | `router.go:279` |

**Fix for each:** Add `get:` operation with:
- `security: [{bearerAuth: []}]`
- `parameters` for path param `id` (uuid format) + query params per handler
- `responses: 200, 401, 403, 404`

### 15. PPID Schema Missing `thumbnail_url`

**File:** `openapi.yaml:55-81`, `domain/ppid/ppid.go:21`

**Issue:** PPID schema doesn't include `thumbnail_url` field, but it exists in the Go entity and is returned in responses.

**Fix:** Add `thumbnail_url: string (format: uri, nullable)` to PPID schema.

### 16. Berita POST/PUT Missing Required `category`

**File:** `openapi.yaml:476-500, 503-533`, `handler/berita.go:55-58, 241-280`

**Issue:** `category` field is documented but not in `required` list.

**Fix:** Add `category` to `required` array. Note: PUT treats all fields as optional (falls back to existing), so for PUT consider documenting partial-update behavior.

### 17. UMKM POST/PUT Missing Required `category` + Wrong Type

**File:** `openapi.yaml:733-768, 770-806`, `handler/umkm.go:54-57, 262-265`

**Issue:** `category` not in `required`, and documented as plain string (should be UUID).

**Fix:** Add to required, document as `format: uuid`.

### 18. PPID POST Missing Required `document` File

**File:** `openapi.yaml:1362-1414`, `handler/ppid.go:106-130`

**Issue:** Document upload is required but not in `required` list.

**Fix:** Add `document` (type: string, format: binary) to `required`.

### 19. PPID PUT Title Not Documented as Required

**File:** `openapi.yaml:1433-1473`, `handler/ppid.go:319`

**Issue:** `title` is required by handler but documented as optional.

**Fix:** Add `title` to `required`.

### 20. PPID Requests List Missing `page`, `limit` Parameters

**File:** `openapi.yaml:1507-1538`, `handler/ppid.go:519-523`

**Issue:** Handler accepts pagination params but OpenAPI doesn't document them.

**Fix:** Add `page` (default 1) and `limit` (default 10, max 100) as query params.

### 21. PPID Download Endpoint Security Wrong

**File:** `openapi.yaml:1622`, `router.go:109`

**Issue:** Endpoint documented with `bearerAuth` security, but route is in PUBLIC group. Token comes from query param `?token=`.

**Fix:** Remove `bearerAuth`, add description clarifying dual-token mode (document token from email OR admin user JWT).

### 22. PPID List Missing `q` (Search) Parameter

**File:** `openapi.yaml:1337`, `handler/ppid.go:209-267`

**Fix:** Add `q` query parameter.

### 23. Struktur POST/PUT `position` Required (Wrong)

**File:** `openapi.yaml:1693-1748`, `handler/struktur.go:46-118, 229-247`

**Issue:** `position` is documented as required but is optional in Go (`*string`).

**Fix:** Move `position` from `required` to optional.

### 24. Profile POST/PUT Body Schema Wrong

**File:** `openapi.yaml:1871-1937`, `handler/profile.go:42-85, 181-231`

**Issue:** OpenAPI requires `[section_name, title, content]`; handler requires `[content, section_name, section_endpoint]`. `title` doesn't exist in Go.

**Fix:**
```yaml
required: [content, section_name, section_endpoint]
properties:
  content: string
    maxLength: 10000
  section_name: string
    maxLength: 100
  section_endpoint: string
    maxLength: 255
    pattern: '^/.*'   # must start with /
  state: boolean
```

### 25. Profile List Missing `section_name`, `state` Filters

**File:** `openapi.yaml:1853-1870`, `handler/profile.go:117-177`

**Fix:** Add `section_name` (string) and `state` (boolean) query params.

### 26. Infographic POST/PUT Body Schema Totally Wrong

**File:** `openapi.yaml:2019-2106`, `handler/infographic.go:44-94, 201-256`

**Issue:** Documents nested `payload.resource.dashboard`; handler expects flat.

**Fix:**
```yaml
required: [component_id, component_type, section_name, section_endpoint]
properties:
  component_id: integer
    minimum: 1
  component_type: string
    enum: [question, dashboard]
  section_name: string
    maxLength: 100
  section_endpoint: string
    maxLength: 255
    pattern: '^/.*'
  category: string
    format: uuid
    nullable: true
  state: boolean
```

### 27. Infographic Preview Token Body Wrong

**File:** `openapi.yaml:2133-2162`, `handler/infographic.go:291-329`

**Issue:** Documents nested `payload.resource.dashboard`; handler expects `component_id, component_type`.

**Fix:** Replace with flat `{component_id: integer, component_type: string enum}`.

### 28. Infographic List Missing `section_name`, `state` Filters

**File:** `openapi.yaml:2001-2018`, `handler/infographic.go:136-197`

**Fix:** Add both query params.

### 29. Desa PUT Body Schema Wrong

**File:** `openapi.yaml:1788-1816`, `handler/desa.go:61-108`

**Issue:** Documents `logo` field that doesn't exist; missing `website` and `vision_mission`.

**Fix:**
```yaml
required: [name]
properties:
  name: string
    maxLength: 255
  description: string
    nullable: true
  address: string
    nullable: true
  phone: string
    maxLength: 50
    nullable: true
  email: string
    format: email
    nullable: true
  website: string
    format: uri
    nullable: true
  vision_mission: string
    nullable: true
```

### 30. Role POST Body Wrong (nested)

**File:** `openapi.yaml:2334-2356`, `handler/role.go:69-126`

**Issue:** Documents `{payload: {role}}`; handler expects flat `{name, permissions}`.

**Fix:**
```yaml
required: [name]
properties:
  name: string
  permissions:
    type: array
    items:
      type: object
      required: [resource, action]
      properties:
        resource: string
        action: string
```

### 31. Role Permissions POST Body Wrong

**File:** `openapi.yaml:2388-2417`, `handler/role.go:220-260`

**Issue:** Same nested bug.

**Fix:** Use flat `{resource: string, action: string}`.

### 32. Role Permissions DELETE — Body vs Path

**File:** `openapi.yaml:2418-2447`, `handler/role.go:266-302`

**Issue:** OpenAPI documents DELETE with request body; handler actually parses `permission` from URL path.

**Fix:** Remove `requestBody`; add `permission` path parameter (format: `resource:action`).

### 33. User-Role POST Body Wrong

**File:** `openapi.yaml:2459-2488`, `handler/role.go:336-370`

**Issue:** Documents nested `{payload: {role}}`; handler expects flat `{role}`.

**Fix:** Use flat schema.

### 34. `/uploads/*` Static Route Not Documented

**File:** `router.go:76`

**Issue:** Public static file server at `/uploads/*` not in OpenAPI.

**Fix:** Add path:
```yaml
/uploads/{filename}:
  get:
    summary: Serve uploaded file
    security: []   # public, no auth
    parameters: [...]
    responses:
      200: { description: File content }
      404: { description: Not found }
```

### 35. Many Endpoints Missing `401`, `403`, `404`, `409` Responses

**Issue:** Most documented operations only have `200` and `400` responses. Missing:
- `401 Unauthorized` — for protected routes
- `403 Forbidden` — for RBAC failures
- `404 Not Found` — for resource lookups
- `409 Conflict` — for duplicate keys, category in use

**Fix:** Add `$ref` responses:
```yaml
responses:
  401: { $ref: '#/components/responses/Unauthorized' }
  403: { $ref: '#/components/responses/Forbidden' }
  404: { $ref: '#/components/responses/NotFound' }
  409: { $ref: '#/components/responses/Conflict' }
```

Define the components:
```yaml
components:
  responses:
    Unauthorized:
      description: Missing or invalid authentication
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    Forbidden:
      description: Insufficient permissions
    NotFound:
      description: Resource not found
    Conflict:
      description: Resource conflict (duplicate, in use)
```

---

## 🟡 MINOR Issues (14)

### 36. Users POST/PUT Multipart Missing in OpenAPI

**File:** `openapi.yaml POST/PUT /users`, `handler/user.go:92-159, 304-383`

**Issue:** Handlers accept multipart with optional `profile_image` file; OpenAPI only documents JSON.

**Fix:** Document both `application/json` and `multipart/form-data` content types.

### 37. Banner POST `status` Field Should Not Exist

**File:** `openapi.yaml:316-343`, `handler/banner.go:46-129`

**Issue:** `status` documented for POST; handler doesn't accept it on create (set via PATCH later).

**Fix:** Remove `status` from POST schema.

### 38. Banner GET Missing `status` Query Filter

**File:** `openapi.yaml:298-315`, `handler/banner.go:135-196`

**Fix:** Add `status` query param (enum: active/inactive).

### 39. Banner POST/PUT Missing `metadata` Field

**File:** `openapi.yaml:316-343, 361-394`, `handler/banner.go:53, 290`

**Fix:** Add `metadata: string (description: JSON-encoded object)`.

### 40. Public File Endpoint Missing Error Responses

**File:** `openapi.yaml:145`, `handler/file.go:30-95`

**Issue:** Documents only 200; handler returns 400/403/404/500.

**Fix:** Add error responses.

### 41. Banners Active List Response Shape Not Documented

**File:** `openapi.yaml:255-261`, `handler/banner.go`

**Issue:** `GET /api/v1/banners/active` documented without pagination/response array.

**Fix:** Document 200 response as array of BannerResponse, with note that limit is hardcoded to 100.

### 42. Profile Sections Names Response Shape Not Documented

**File:** `openapi.yaml:1888+`, `handler/profile.go:259-261`

**Issue:** Handler returns `{section_names: [...]}` but shape not described.

**Fix:** Add response schema.

### 43. PPID Categories POST/DELETE Missing `bearerAuth`

**File:** `openapi.yaml:1145-1200`, `router.go:271`

**Fix:** Add `security: [{bearerAuth: []}]` to both.

### 44. Infographic Detail Missing `token` Field

**File:** `handler/infographic.go:128`

**Issue:** Authenticated GET returns `token` field but it's not documented.

**Fix:** Add `token: string` to authenticated InfographicResponse schema.

### 45. PPID Document URL Description Misleading

**File:** `openapi.yaml PPID schema`, `handler/ppid.go:194, 297, 414`

**Issue:** Description says "URL to download document (for approved requests only)" — but URL is always returned regardless of approval status.

**Fix:** Update description to remove "approved requests only" qualifier.

### 46. Login Response Shape Not Documented

**File:** `handler/auth.go:Login`

**Fix:** Document 200 response:
```yaml
schema:
  type: object
  required: [access_token, refresh_token, expires_at]
  properties:
    access_token: string
    refresh_token: string
    expires_at: string
      format: date-time
```

### 47-53. Other minor inconsistencies

- Pagination default values (page=1, limit=10) not consistently documented
- Some timestamps documented as `date-time` others as `string`
- Some `*string` nullable fields not marked `nullable: true`
- Missing `x-required-permissions` extension for RBAC scope documentation

---

## Missing Component Schemas (17+)

Only 4 schemas exist. Need to add:

| Schema | Status | Priority |
|---|---|---|
| `UserResponse` | Missing | 🟠 MAJOR |
| `RoleResponse` | Missing | 🟠 MAJOR |
| `PermissionResponse` | Missing | 🟠 MAJOR |
| `DesaResponse` | Missing | 🟠 MAJOR |
| `BeritaResponse` | Missing | 🟠 MAJOR |
| `BeritaListResponse` | Missing | 🟠 MAJOR |
| `BeritaCategoryResponse` | Missing | 🟠 MAJOR |
| `BannerResponse` | Missing | 🟠 MAJOR |
| `UMKMResponse` | Missing | 🟠 MAJOR |
| `UMKMCategoryResponse` | Missing | 🟠 MAJOR |
| `FasilitasResponse` | Missing | 🟠 MAJOR |
| `FasilitasCategoryResponse` | Missing | 🟠 MAJOR |
| `StrukturResponse` | Missing | 🟠 MAJOR |
| `ProfileResponse` | Missing | 🟠 MAJOR |
| `ProfileSectionNamesResponse` | Missing | 🟢 MINOR |
| `InfographicResponse` | Missing | 🟠 MAJOR |
| `InfographicCategoryResponse` | Missing | 🟠 MAJOR |
| `PPIDResponseTruncated` | Missing | 🟡 MINOR |
| `PPIDResponsePrivate` | Missing | 🟡 MINOR |

---

## Recommendations

### Process Improvements

1. **Adopt single source of truth for schemas**
   - Generate Go structs from OpenAPI via `oapi-codegen` or `ogen`
   - OR generate OpenAPI from Go structs via `swaggo/swag`
   - Eliminates hand-maintained drift

2. **Add OpenAPI validation to CI**
   ```bash
   npx @redocly/cli lint openapi.yaml
   ```
   - Would catch invalid paths like item #1
   - Run on every PR

3. **Endpoint parity test**
   - Parse `router.go` AST
   - Parse `openapi.yaml`
   - Diff path/method sets
   - Fail CI on drift

4. **Schema validation tests**
   - Use `schemathesis` to fuzz handlers against documented schema
   - Assert every response matches

### Quick-Win Fixes (Low Effort, ~1 day total)

1. Remove `/api/v1/public/ppid/categories (legacy)` from `openapi.yaml:1498`
2. Add `/uploads/*` path to OpenAPI
3. Fix 4 path typos (berita, umkm, ppid categories)
4. Add `bearerAuth` to PPID categories POST/DELETE
5. Add `category` to berita/umkm required lists
6. Add `document` to PPID POST required
7. Fix 4 wrong request body shapes (desa, profile, infographic, role)
8. Fix integration test paths for PPID download

### Medium-Effort Fixes (~1 week)

1. Add all 17 missing component schemas
2. Add `401`, `403`, `404`, `409` responses to all protected endpoints
3. Add `x-required-permissions` extension documenting RBAC scope
4. Generate schemas from Go structs via tooling

### Architectural Improvements (~2-3 weeks)

1. Switch to `oapi-codegen` for type-safe API server/client generation
2. Add `make openapi-verify` target in Makefile
3. Add CI step: `schemathesis` for fuzzing
4. Document RBAC scopes per endpoint via custom extension
5. Consider adopting `ogen` for full code generation from spec

---

## File Locations Summary

| Concern | File |
|---|---|
| OpenAPI spec | `openapi.yaml` |
| Router (routes) | `interface/http/router.go` |
| HTTP handlers | `interface/http/handler/*.go` |
| Domain entities (response shapes) | `domain/*/*.go` |
| Use case response types | `usecase/*/*.go` |
| Existing partial verification | `openapi-verification.md` |

---

## Related Documentation

- [Bugs — backup routes commented](./bugs.md#4-backup-http-routes-commented-out)
- [Architecture — OpenAPI drift](./architecture.md#3-no-specification-layer-openapi-as-single-source-of-truth)
- [Code quality — long handler files](./code-quality.md#4-long-handler-files-500-lines)