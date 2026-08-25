
## Update — Task 7.4 (signed URL rollout)

**As of the 7.4 commit (signed URL foundation complete):**

- The 10 per-scope public media routes (`/api/v1/public/gallery/media/{id}/...`, `/api/v1/public/{banner,berita,struktur}/media/{id}/...`, `/api/v1/gallery/media/{id}/...`) are **gone**.
- The legacy PPID download route `/api/v1/ppid/document/{id}/download?token=...` is **gone** (replaced by the unified signed-URL flow with the JWT embedded in the approval email).
- Every media URL the server returns now goes through `GET /api/v1/media/{id}/content?jwt=...` and `GET /api/v1/media/{id}/thumbnail?jwt=...`.
- The single new endpoint `POST /api/v1/media/refresh?jwt=<old>&id=<media-id>` lets clients transparently renew expiring tokens (admin URLs are 1h).
- For an admin session, the URL `media.url` carries `scope=admin` (1h TTL); for anonymous visitors, `scope=public` (24h TTL). The same URL shape is used in every response.

**Client integration change** is just URL handling:

```tsx
// ✅ current (post-7.4): render the URL the server returned
<img src={banner.media.thumbnail_url ?? banner.media.url} />

// ❌ old (pre-7.4): the URL shape used to vary by feature
const url = `${API}/api/v1/public/banner/media/${id}/content`;  // was feature-specific
const url2 = `${API}/api/v1/gallery/media/${id}/content`;         // was admin-specific
```

The `media.url` field in every response is now always `/api/v1/media/{id}/content?jwt=<signed>`. The old per-scope URLs no longer exist; any bookmarked pre-7.4 link will 404.

## PPID upload — combined endpoint (document + thumbnail)

`POST /api/v1/ppid/upload` accepts both files in one multipart request
and returns both media ids, so the admin client needs only one upload
round-trip before creating the PPID record:

```http
POST /api/v1/ppid/upload
Authorization: Bearer <admin-jwt>
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="document"; filename="peraturan.pdf"
Content-Type: application/pdf
<binary>
--boundary
Content-Disposition: form-data; name="thumbnail"; filename="thumb.webp"
Content-Type: image/webp
<binary>
--boundary--
```

```json
{
  "success": true,
  "data": {
    "document":  { "url": "...", "media_id": "...", "filename": "peraturan.pdf" },
    "thumbnail": { "url": "...", "media_id": "...", "filename": "thumb.webp" }
  }
}
```

Rules:
- `document` and `thumbnail` are both optional, but **at least one is required** (else 400).
- Each file is validated against its own allowlist (document → PDF/DOC/DOCX/XLS/XLSX; thumbnail → image).
- Attach the returned ids via `document_media_id` / `thumbnail_media_id` on `POST /ppid` or `PUT /ppid/{id}`.

The single-file endpoints `POST /ppid/upload-media` (document) and
`POST /ppid/upload-thumbnail` (thumbnail) remain as aliases for
existing clients.

```typescript
const fd = new FormData();
fd.append('document', docFile);
if (thumbFile) fd.append('thumbnail', thumbFile);
const { data } = await api.post('/api/v1/ppid/upload', fd, {
  headers: { 'Content-Type': 'multipart/form-data' },
});
// data.document.media_id, data.thumbnail.media_id
```
