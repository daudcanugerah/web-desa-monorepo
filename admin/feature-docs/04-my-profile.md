# 04 — My Profile

Self-service profile page for the currently logged-in user. View + edit identity, change own password.

---

## 1. Purpose

Let any authenticated user:
- View their profile (avatar, name, email, roles).
- Update their name and avatar.
- Change their own password (requires the old password).

This is the non-admin counterpart to `03-user-management.md`.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/me` | `MyProfile` | Profil Saya | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/users/MyProfileView.vue` | Profile view + edit + password modal |
| `src/services/user.service.js` | Reused for `changePassword` |
| `src/components/forms/ImageUpload.vue` | Avatar upload |
| `src/components/common/AppBadge.vue` | Role pill |
| `src/components/common/AppModal.vue` | Password-change modal |
| `src/components/common/AppIconButton.vue` | Action buttons |

Note: `MyProfileView` calls `api.put('/users/me', ...)` directly for the profile update (not via a service). The password change goes through `userService.changePassword`.

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET` | `/users/me` | — | Implicitly via `authStore.fetchCurrentUser` (header avatar) |
| `PUT` | `/users/me` | `{ name?, avatar? }` (multipart when avatar) | MyProfileView (save) |
| `PUT` | `/users/{id}/password` | `{ old_password, new_password }` | MyProfileView (password modal) |

For the password change the `{id}` is `currentUser.id`.

---

## 5. Data Model

```ts
type CurrentUser = {
  id: string
  email: string
  name: string
  avatar?: string | null
  avatar_url?: string | null
  roles: string[]
}
```

---

## 6. Behaviors

- On mount, reads from `authStore.currentUser`. If absent, the layout's mount should have already fetched it.
- **Edit profile**: toggles inline edit mode; avatar preview via `useImagePreview`; on save sends `PUT /users/me` (multipart when avatar present, JSON otherwise).
- **Change password**:
  - Opens an `AppModal`.
  - Fields: `old_password`, `new_password`, `new_password_confirmation`.
  - Validation: old required, new ≥ 8 chars, confirmation matches.
  - On submit → `userService.changePassword(currentUser.id, { old_password, new_password })`.
- After successful profile save → refreshes `authStore.currentUser` via `fetchCurrentUser()` so the header reflects new info.

---

## 7. UI Notes

- Top section: avatar (96x96 round) + name + email + roles badges.
- Edit form appears inline when toggled.
- Password modal is centered, with three password inputs.

---

## 8. Notes & Gotchas

- **`PUT /users/me`** is called directly with the raw `api` instance, bypassing any service abstraction. Future feature should be moved to `userService.updateMe` for consistency.
- **Avatar upload** re-sends the file via `FormData`. If the user does not change the avatar, the request is JSON — verify the backend accepts both shapes.
- **No email change.** Email is read-only on this page (admins can change it via `/users/:id/edit`).
- **No "delete my account"** flow.