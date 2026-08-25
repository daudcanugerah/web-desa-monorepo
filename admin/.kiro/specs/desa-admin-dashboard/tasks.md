# Implementation Plan: Desa Admin Dashboard

## Delivery Strategy

Tasks are organized as **user stories** — each story delivers a working, testable slice of the app.
Complete one story, verify it in the browser, give feedback, then move to the next.

---

## Story 1: "I can open the app in a browser"

> As a developer, I want a running Vue 3 app with routing so I have a working foundation.

- [x] 1. Scaffold project and configure tooling
  - [x] 1.1 Scaffold Vite + Vue 3 project (`npm create vite@latest desa-admin -- --template vue`)
  - [x] 1.2 Install core deps: `vue-router@4`, `pinia`, `axios`, `tailwindcss`, `postcss`, `autoprefixer`
  - [x] 1.3 Configure `tailwind.config.js`, `postcss.config.js`, and `src/assets/styles/main.css` with Tailwind directives
  - [x] 1.4 Configure `vite.config.js` with `@vitejs/plugin-vue` and `@/` path alias
  - [x] 1.5 Create `.env` and `.env.example` with `VITE_API_BASE_URL=http://localhost:8080/api/v1`
  - [x] 1.6 Create `.gitignore` (node_modules, dist, .env, .env.local, *.log, .DS_Store)
  - [x] 1.7 Set up `src/main.js` — mount app with Pinia and Vue Router
  - [x] 1.8 Create minimal `src/App.vue` with `<router-view />`
  - _Requirements: 25.1–25.8, 27.1–27.2_

- [x] 2. Router skeleton with placeholder pages
  - [x] 2.1 Create `src/router/index.js` with routes: `/login`, `/` (dashboard), `/:pathMatch(.*)` (404)
  - [x] 2.2 Create stub views: `LoginView.vue`, `DashboardView.vue`, `NotFoundView.vue` — each just renders its name
  - [x] 2.3 Add `router.beforeEach` guard — redirect unauthenticated users to `/login` (read token from `localStorage`)
  - _Requirements: 15.3, 15.4, 15.5, 15.7_

**✅ Checkpoint**: `npm run dev` opens the app, `/login` loads, navigating to `/` redirects to `/login`.

- Commit: `chore: scaffold Vue 3 + Vite + Tailwind + router skeleton`

---

## Story 2: "I can log in and see the app shell"

> As an admin, I want to log in with my email and password so I can access the dashboard.

- [x] 3. Auth infrastructure
  - [x] 3.1 Create `src/utils/storage.js` — `getToken`, `setTokens`, `clearTokens` using `localStorage`
  - [x] 3.2 Create `src/stores/auth.js` — state: `accessToken`, `refreshToken`, `expiresAt`, `currentUser`; actions: `setTokens`, `clearTokens`, `fetchCurrentUser` (calls `GET /users/me`)
  - [x] 3.3 Create `src/services/api.js` — Axios instance with `baseURL`, 30s timeout; request interceptor attaches Bearer token; response interceptor handles 401 → refresh → retry → redirect to `/login` on failure
  - [x] 3.4 Create `src/services/auth.service.js` — `login(email, password)`, `refresh(refreshToken)`
  - [x] 3.5 Update router guard to read `accessToken` from auth store (not raw localStorage)
  - _Requirements: 1.2, 1.3, 2.1, 2.4, 2.5, 22.1–22.6_

- [x] 4. Login view
  - [x] 4.1 Create `src/components/common/AppButton.vue` — primary/secondary/danger variants, loading state, disabled
  - [x] 4.2 Create `src/components/common/AppInput.vue` — label, error message, accessible aria attributes
  - [x] 4.3 Build `LoginView.vue` — email + password fields, client-side required validation, calls `authService.login`, stores tokens, redirects to `/`, shows API error message
  - _Requirements: 1.1, 1.4, 1.5, 1.6, 14.1, 14.5, 14.7_

- [x] 5. App shell layout
  - [x] 5.1 Create `src/stores/notification.js` — toast queue; success auto-dismisses (3s), errors persist
  - [x] 5.2 Create `src/components/common/AppNotification.vue` — renders toast stack, dismiss button on errors
  - [x] 5.3 Create `src/components/layout/AppSidebar.vue` — nav links (stub routes ok), active highlight, collapsible on mobile
  - [x] 5.4 Create `src/components/layout/AppHeader.vue` — shows current user name, logout button (clears tokens → `/login`)
  - [x] 5.5 Create `src/components/layout/AppLayout.vue` — sidebar + header + `<router-view>`, responsive Tailwind grid
  - [x] 5.6 Wrap all authenticated routes with `AppLayout` in the router
  - [x] 5.7 Update `App.vue` to render `<AppNotification>` globally
  - _Requirements: 2.3, 13.1–13.5, 15.1, 15.2, 20.1–20.6_

**✅ Checkpoint**: Log in → redirected to dashboard shell with sidebar and header. Logout works. Wrong credentials shows error. Toast notifications appear.

- Commit: `feat(auth): implement login flow, app shell layout, and toast notifications`

---

## Story 3: "I can see the dashboard overview"

> As an admin, I want to see key stats on the home page so I know the system status at a glance.

- [x] 6. Dashboard stats page
  - [x] 6.1 Create `src/components/common/LoadingSpinner.vue`
  - [x] 6.2 Create `src/views/dashboard/DashboardView.vue` — four stat cards (Users, Berita, UMKM, Fasilitas); fetch counts in parallel from `pagination.total` of each list endpoint (`?limit=1`); loading skeleton; error notification on failure
  - [x] 6.3 Add breadcrumb meta to dashboard route; create `src/components/layout/AppBreadcrumb.vue`
  - _Requirements: 3.1–3.6, 15.6, 23.1_

**✅ Checkpoint**: Dashboard shows 4 stat cards with real counts from the API. Loading spinner visible during fetch.

- Commit: `feat(dashboard): implement dashboard overview with stat cards`

---

## Story 4: "I can manage users"

> As an admin, I want to list, create, edit, and delete user accounts.

- [x] 7. Shared list/table components
  - [x] 7.1 Create `src/components/common/AppTable.vue` — columns config + data array, loading skeleton, empty state
  - [x] 7.2 Create `src/components/common/AppPagination.vue` — page numbers, prev/next, driven by `{ page, total_pages }`
  - [x] 7.3 Create `src/components/common/AppModal.vue` — backdrop, ESC close, focus trap
  - [x] 7.4 Create `src/components/common/ConfirmDialog.vue` — uses AppModal, driven by confirm store state
  - [x] 7.5 Create `src/stores/ui.js` — `confirmDialog` state; `src/composables/useConfirm.js` — Promise-based confirm
  - [x] 7.6 Update `App.vue` to render `<ConfirmDialog>` globally
  - _Requirements: 16.1–16.6, 19.1–19.5, 23.1–23.2_

- [x] 8. User management views
  - [x] 8.1 Create `src/services/user.service.js` — `list`, `get`, `create`, `update`, `delete`, `updatePassword`, `assignRole`, `removeRole`
  - [x] 8.2 Create `src/views/users/UserListView.vue` — AppTable (name/email/roles columns), search with 300ms debounce, AppPagination, delete with ConfirmDialog
  - [x] 8.3 Create `src/views/users/UserFormView.vue` — create: name/email/password; edit: name/email + separate password change section; role assignment dropdown
  - [x] 8.4 Add `/users` and `/users/create`, `/users/:id/edit` routes to router
  - _Requirements: 4.1–4.9, 17.1–17.4_

**✅ Checkpoint**: Navigate to Users → see paginated list, search works, create a user, edit a user, delete with confirmation dialog.

- Commit: `feat(users): implement user management (list, create, edit, delete, password change)`

---

## Story 5: "I can manage roles and permissions"

> As an admin, I want to create roles and assign permissions so I can control access.

- [x] 9. Role management
  - [x] 9.1 Create `src/services/role.service.js` — `list`, `create`, `delete`, `getPermissions`, `addPermission`, `removePermission`
  - [x] 9.2 Create `src/views/users/RoleListView.vue` — list roles, expandable permissions per role (checkboxes for known resource:action pairs), create role inline, delete with ConfirmDialog
  - [x] 9.3 Add `/roles` route
  - _Requirements: 5.1–5.9_

**✅ Checkpoint**: Navigate to Roles → create a role, add/remove permissions, delete a role.

- Commit: `feat(roles): implement role and permission management`

---

## Story 6: "I can manage banners"

> As an admin, I want to upload and manage promotional banners shown on the public site.

- [x] 10. Banner management
  - [x] 10.1 Create `src/utils/validators.js` — MIME type check, file size helpers
  - [x] 10.2 Create `src/composables/useImagePreview.js` — preview URL, file size display, MIME validation
  - [x] 10.3 Create `src/components/forms/ImageUpload.vue` — drag-drop, preview, warns >5MB, rejects >10MB, accepts JPEG/PNG/WebP
  - [x] 10.4 Create `src/services/banner.service.js` — `list`, `get`, `create`, `update`, `delete`, `updateStatus`
  - [x] 10.5 Create `src/views/content/BannerListView.vue` — AppTable (thumbnail/title/status), status filter, inline status toggle, delete with ConfirmDialog
  - [x] 10.6 Create `src/views/content/BannerFormView.vue` — title + ImageUpload + status toggle; multipart submit
  - [x] 10.7 Add `/banners` routes
  - _Requirements: 6.1–6.10, 18.1–18.6_

**✅ Checkpoint**: Upload a banner image with preview, see it in the list, toggle active/inactive, delete it.

- Commit: `feat(banners): implement banner management with image upload`

---

## Story 7: "I can manage news articles"

> As an admin, I want to write and publish news articles with a rich text editor.

- [x] 11. News management
  - [x] 11.1 Install `@tiptap/vue-3` and `@tiptap/starter-kit`
  - [x] 11.2 Create `src/components/forms/RichTextEditor.vue` — TipTap with bold/italic/lists/headings toolbar
  - [x] 11.3 Create `src/services/news.service.js` — `list`, `get`, `create`, `update`, `delete`
  - [x] 11.4 Create `src/views/content/NewsListView.vue` — AppTable (title/date), search + date range filter, AppPagination, delete
  - [x] 11.5 Create `src/views/content/NewsFormView.vue` — title, RichTextEditor, ImageUpload, multipart submit
  - [x] 11.6 Add `/berita` routes
  - _Requirements: 7.1–7.11, 17.1–17.3_

**✅ Checkpoint**: Write an article with formatted content, attach an image, save it, see it in the list.

- Commit: `feat(berita): implement news article management with rich text editor`

---

## Story 8: "I can manage UMKM listings"

> As an admin, I want to add and manage small business listings.

- [x] 12. UMKM management
  - [x] 12.1 Create `src/services/umkm.service.js` — `list`, `get`, `create`, `update`, `delete`
  - [x] 12.2 Create `src/views/content/UmkmListView.vue` — AppTable (name/owner/phone), search, AppPagination, delete
  - [x] 12.3 Create `src/views/content/UmkmFormView.vue` — name, owner, address, phone, email, website; JSON body; phone validation
  - [x] 12.4 Add `/umkm` routes
  - _Requirements: 8.1–8.9, 17.1–17.3_

**✅ Checkpoint**: Add a business listing, edit it, search by name, delete it.

- Commit: `feat(umkm): implement UMKM business listing management`

---

## Story 9: "I can manage facilities with a map"

> As an admin, I want to add village facilities and pin their location on a map.

- [x] 13. Fasilitas management
  - [x] 13.1 Install `leaflet`
  - [x] 13.2 Create `src/components/forms/MapPicker.vue` — Leaflet map, click to set marker, emits lat/lng, syncs with input fields
  - [x] 13.3 Create `src/services/facility.service.js` — `list`, `get`, `create`, `update`, `delete`
  - [x] 13.4 Create `src/views/content/FacilityListView.vue` — AppTable (name/type/coordinates), AppPagination, delete
  - [x] 13.5 Create `src/views/content/FacilityFormView.vue` — name, type, description, lat/lng inputs + MapPicker; Yup validation lat(-90..90) lng(-180..180)
  - [x] 13.6 Add `/fasilitas` routes
  - _Requirements: 9.1–9.12_

**✅ Checkpoint**: Add a facility, click the map to set coordinates, see lat/lng fields update, save and see it in the list.

- Commit: `feat(fasilitas): implement facility management with Leaflet map picker`

---

## Story 10: "I can manage PPID documents"

> As an admin, I want to upload public information disclosure documents.

- [x] 14. PPID management
  - [x] 14.1 Create `src/components/forms/FileUpload.vue` — accepts PDF/DOC/DOCX/XLS/XLSX, max 50MB, shows filename/size
  - [x] 14.2 Create `src/services/ppid.service.js` — `list`, `get`, `create`, `update`, `delete`, `listRequests`
  - [x] 14.3 Create `src/views/content/PpidListView.vue` — AppTable (title/category/date), category filter, search, AppPagination, delete; section for public requests
  - [x] 14.4 Create `src/views/content/PpidFormView.vue` — title, category, description, FileUpload; multipart submit
  - [x] 14.5 Add `/ppid` routes
  - _Requirements: 10.1–10.10, 17.5–17.6_

**✅ Checkpoint**: Upload a PDF document, see it in the list with category filter, delete it.

- Commit: `feat(ppid): implement PPID document management with file upload`

---

## Story 11: "I can manage the organizational structure"

> As an admin, I want to manage staff members and their positions.

- [x] 15. Struktur management
  - [x] 15.1 Create `src/services/structure.service.js` — `list`, `get`, `create`, `update`, `delete`
  - [x] 15.2 Create `src/views/organization/StructureView.vue` — AppTable (name/position/photo), search, AppPagination; create/edit in AppModal with ImageUpload; delete with ConfirmDialog
  - [x] 15.3 Add `/struktur` route
  - _Requirements: 11.1–11.10_

**✅ Checkpoint**: Add a staff member with a photo, edit their position, delete them.

- Commit: `feat(struktur): implement organizational structure management`

---

## Story 12: "I can manage the village profile"

> As an admin, I want to update the village's general information.

- [x] 16. Village profile
  - [x] 16.1 Create `src/services/profile.service.js` — `get()`, `update(data)`
  - [x] 16.2 Create `src/views/organization/ProfileView.vue` — display profile fields (name, address, phone, email, website, vision_mission); edit form; JSON PUT; success notification
  - [x] 16.3 Add `/profile` route
  - _Requirements: 12.1–12.6_

**✅ Checkpoint**: View village profile, edit and save, see success notification.

- Commit: `feat(desa): implement village profile view and edit`

---

## Story 13: "I can reset my password"

> As an admin, I want to recover my account if I forget my password.

- [x] 17. Password reset flow
  - [x] 17.1 Create `src/services/auth.service.js` additions — `requestPasswordReset(email)`, `checkResetToken(token)`, `confirmPasswordReset(token, newPassword)`
  - [x] 17.2 Create `src/views/auth/ForgotPasswordView.vue` — email form, shows confirmation on success
  - [x] 17.3 Create `src/views/auth/ResetPasswordView.vue` — reads `?token=` from query, validates token on mount, new password + confirm form, redirects to `/login` on success
  - [x] 17.4 Add "Forgot Password" link to LoginView
  - _Requirements: 21.1–21.9_

**✅ Checkpoint**: Click "Forgot Password" on login page, submit email, follow reset link, set new password, redirected to login.

- Commit: `feat(auth): implement forgot password and reset password flow`

---

## Story 14: "Final wiring and cleanup"

- [x] 18. Final integration
  - [x] 18.1 Add all remaining sidebar nav links in AppSidebar pointing to real routes
  - [x] 18.2 Add `meta.breadcrumb` labels to all routes; verify AppBreadcrumb renders correctly on all pages
  - [x] 18.3 Add `NotFoundView` catch-all route
  - [x] 18.4 Verify consistent naming conventions across all files (Req 26.8)
  - [x] 18.5 Run `npm run build` — fix any build errors
  - _Requirements: 15.6, 26.1–26.8_

**✅ Checkpoint**: Full build passes. All nav links work. Breadcrumbs show on every page.

- Commit: `chore: final wiring, breadcrumbs, sidebar links, and build verification`

---

## Story 15: "Feature refinements and improvements"

- [x] 19. Role and user management enhancements
  - [x] 19.1 Prevent deletion of default roles (admin, operator)
  - [x] 19.2 Display created_at timestamp in role list
  - [x] 19.3 Remove permission button from role actions (keep edit and delete only)
  - [x] 19.4 Add created_at column to user list table
  - _Requirements: 4.1–4.9, 5.1–5.9_

- [x] 20. News article enhancements
  - [x] 20.1 Implement full Vue Quill toolbar with formatting options (bold, italic, underline, strikethrough, code, headings, lists, blockquote, code block, horizontal rule, clear format)
  - [x] 20.2 Add image upload button to rich text editor toolbar
  - [x] 20.3 Add video upload button to rich text editor toolbar
  - [x] 20.4 Integrate media upload endpoint (`/berita/upload-media`) for images and videos
  - _Requirements: 7.1–7.11_

- [x] 21. Upload process improvements
  - [x] 21.1 Display uploaded images using endpoint `<base-url-api>/api/v1/files/<img-name>` after upload
  - [x] 21.2 Show image preview when editing forms
  - [x] 21.3 Ensure consistent image display across all content management features
  - [x] 21.4 Create centralized image URL utility for consistent API endpoint usage
  - _Requirements: 6.1–6.10, 18.1–18.6_

- [x] 22. User profile enhancements
  - [x] 22.1 Create dedicated user profile view page (separate from edit)
  - [x] 22.2 Display user information on profile page
  - [x] 22.3 Add edit profile button to navigate to edit form
  - [x] 22.4 Add update password button on profile page
  - [x] 22.5 Integrate with existing MyProfileView edit and password change functionality
  - _Requirements: 4.1–4.9_

**✅ Checkpoint**: All refinements implemented and tested. Roles cannot be deleted, timestamps display correctly, news editor has full toolbar with media upload, profile page shows user info with edit and password buttons.

- Commit: `feat: add role/user enhancements, full news editor toolbar, and profile improvements`

---

## Story 16: "Additional feature enhancements"

- [x] 23. Banner enhancements
  - [x] 23.1 Add link field to banner form
  - [x] 23.2 Display link in banner list table
  - _Requirements: 6.1–6.10_

- [x] 24. Berita category management
  - [x] 24.1 Add category field to news form (create/edit)
  - [x] 24.2 Display category in news list table
  - _Requirements: 7.1–7.11_

- [x] 25. UMKM enhancements
  - [x] 25.1 Add category field to UMKM form (create/edit)
  - [x] 25.2 Display category in UMKM list table
  - [x] 25.3 Add multiple image upload capability
  - _Requirements: 8.1–8.9_

- [x] 26. PPID enhancements
  - [x] 26.1 Add publication_at date field to PPID form
  - [x] 26.2 Display publication_at in PPID list table (YYYY-MM-DD format)
  - [x] 26.3 Create PPID requests menu with list view
  - [x] 26.4 Add approve/reject action buttons with confirmation dialogs
  - _Requirements: 10.1–10.10_

- [x] 27. Fasilitas enhancements
  - [x] 27.1 Add multiple image upload capability
  - [x] 27.2 Store and manage multiple images per facility
  - _Requirements: 9.1–9.12_

---

## Story 17: "Profile section management"

> As an admin, I want to manage profile sections with categories so I can organize village information better.

- [ ] 28. Profile section CRUD
  - [ ] 28.1 Create `src/services/profile-section.service.js` — `list`, `get`, `create`, `update`, `delete`
  - [ ] 28.2 Update sidebar to have Profile submenu with "General Info" and "Profile Sections"
  - [ ] 28.3 Create `src/views/organization/ProfileSectionListView.vue` — AppTable (name/category/order), search, AppPagination, delete with ConfirmDialog
  - [ ] 28.4 Create `src/views/organization/ProfileSectionFormView.vue` — name, category, content (RichTextEditor), order; JSON body
  - [ ] 28.5 Add `/profile/sections` routes for list, create, edit
  - [ ] 28.6 Update existing ProfileView to be under "General Info" submenu
  - _Requirements: Profile section API endpoints from swagger_

**✅ Checkpoint**: Navigate to Profile → Profile Sections, create a new section with category, edit content with rich text editor, see it in the list.

- Commit: `feat(profile): implement profile section management with categories`

---

## Story 18: "Infographic dashboard management (Metabase integration)"

> As an admin, I want to manage Metabase dashboards for infographics with preview functionality.

- [ ] 29. Infographic CRUD with preview
  - [ ] 29.1 Create `src/services/infographic.service.js` — `list`, `get`, `create`, `update`, `delete`, `getPreview(token)`
  - [ ] 29.2 Create `src/views/content/InfographicListView.vue` — AppTable (title/description/status), search, AppPagination, delete with ConfirmDialog
  - [ ] 29.3 Create `src/views/content/InfographicFormView.vue` — title, description, metabase_token, status; preview button that calls `/infographic/preview/{token}` and shows iframe
  - [ ] 29.4 Add automatic preview loading when editing existing infographic
  - [ ] 29.5 Add `/infographic` routes for list, create, edit
  - [ ] 29.6 Add "Infographic" menu item to sidebar
  - _Requirements: Infographic API endpoints from swagger, Metabase integration_

**✅ Checkpoint**: Navigate to Infographic, create a dashboard with Metabase token, click preview to see embedded dashboard, edit existing infographic shows preview automatically.

- Commit: `feat(infographic): implement Metabase dashboard management with preview`

---

- Each story ends with a **✅ Checkpoint** — verify it works before starting the next story
- Stories 1–3 are the critical path — complete these first to have a working skeleton
- Stories can be parallelized after Story 3 if needed
- Dashboard stats use `pagination.total` from list endpoints — no dedicated stats API
- Struktur is a flat list (not a tree) per the API
- UMKM uses JSON body (no file upload)
- Fasilitas uses bounding box filter, not text search
