# Features

## 1. Authentication

### 1.1 Login

- User enters email and password
- System validates credentials against API
- On success: JWT tokens stored, redirect to dashboard
- On failure: display error message
- Auto-redirect to `/` if already authenticated

### 1.2 Forgot Password

- User requests password reset by entering email
- System sends reset instructions to email (API handles sending)

### 1.3 Reset Password

- User sets new password using token from email link
- User confirms new password (must match)

---

## 2. Dashboard

### 2.1 Stats Overview

- Display 4 stat cards: Users, Berita, UMKM, Fasilitas
- Each card shows total count from respective API
- Cards display icon and label with count
- Loading spinner while fetching

---

## 3. User Management

### 3.1 User List (`/users`)

- View paginated table of all users
- Columns: Name, Email, Roles, Dibuat, Actions
- Search users by name/email (debounced 300ms)
- Filter roles by tag display
- Pagination with page numbers
- Click "Tambah User" → navigate to create form

### 3.2 User Create (`/users/create`)

- Fill in: Name (required), Email (required), Password (required, min 8 chars)
- Upload profile photo (optional)
- Select role from available roles (radio buttons)
- Click "Buat User" → create user, redirect to list

### 3.3 User Edit (`/users/:id/edit`)

- Update Name, Email, Profile photo
- Change assigned role
- Click "Perbarui User" → save changes
- Separate "Ubah Password" section:
  - Enter Old Password (optional for admin)
  - Enter New Password (required, min 8 chars)
  - Confirm New Password (must match)
  - Click "Ubah Password" → update password

### 3.4 Delete User

- Click "Hapus" button on user row
- Confirm dialog: "Hapus [name]?"
- Cannot delete own account (button shows "—" for current user)
- Success notification → refresh list

### 3.5 My Profile (`/me`)

- View profile card: Avatar (image or initial), Name, Email, Roles, Bergabung date
- Click "Edit Profil" → enter edit mode
- Click "Ubah Password" → open password modal
- Edit mode:
  - Update Name, Email (disabled), Profile photo
  - Click "Simpan Perubahan" → save

---

## 4. Role Management

### 4.1 View Roles (`/roles`)

- View list of all available roles
- See role names and descriptions

---

## 5. Banner Management

### 5.1 Banner List (`/banners`)

- View paginated table of banners
- Columns: Judul, Link, Status, Dibuat, Actions
- Filter by status: Semua, Aktif, Nonaktif (toggle buttons)
- Toggle status inline: click status badge → toggle active/inactive
- Click "Tambah Banner" → navigate to create form

### 5.2 Banner Create (`/banners/create`)

- Enter Title (required)
- Enter Link URL (optional, e.g., <https://example.com>)
- Upload Image (required, shown as preview)
- Click "Buat Banner" → create banner, redirect to list

### 5.3 Banner Edit (`/banners/:id/edit`)

- Update Title, Link, Image
- Toggle Status (active/inactive toggle switch)
- Click "Simpan Perubahan" → save changes

### 5.4 Delete Banner

- Click "Hapus" → confirm dialog → delete

---

## 6. Berita (News) Management

### 6.1 News List (`/berita`)

- View paginated table of articles
- Columns: Judul, Kategori, Dibuat, Diperbarui, Actions
- Search by title (debounced 300ms)
- Filter by category (dropdown)
- Filter by date range: date-from / date-to
- Click "Tambah Berita" → navigate to create form

### 6.2 News Create (`/berita/create`)

- Enter Judul (required)
- Enter Kategori (optional, e.g., Pengumuman, Berita)
- Write content using Rich Text Editor (Quill)
- Upload Gambar (optional, featured image)
- Click "Publikasikan" → create article, redirect to list

### 6.3 News Edit (`/berita/:id/edit`)

- Update all fields
- Click "Simpan" → save changes

### 6.4 Delete News

- Click "Hapus" → confirm dialog → delete

---

## 7. UMKM (Small Business) Management

### 7.1 UMKM List (`/umkm`)

- View paginated table of businesses
- Columns: Nama Usaha, Pemilik, Kategori, Telepon, Dibuat/Diperbarui, Actions
- Search by nama usaha (debounced 300ms)
- Click "Tambah UMKM" → navigate to create form

### 7.2 UMKM Create (`/umkm/create`)

- Enter Nama Usaha (required)
- Enter Pemilik (required)
- Enter Alamat (required)
- Enter Telepon (required, validated format)
- Enter Email (optional)
- Enter Website (optional)
- Enter Kategori (optional, e.g., Makanan, Kerajinan)
- Enter Deskripsi (optional)
- Upload multiple Gambar (optional)
- Remove uploaded images before saving
- Click "Tambah" → create, redirect to list

### 7.3 UMKM Edit (`/umkm/:id/edit`)

- Update all fields
- Manage images: view existing, remove, add new
- Click "Simpan" → save changes

### 7.4 Delete UMKM

- Click "Hapus" → confirm dialog → delete

---

## 8. Fasilitas (Facility) Management

### 8.1 Facility List (`/fasilitas`)

- View paginated table of facilities
- Columns: Nama, Tipe, Koordinat, Actions
- Click "Tambah Fasilitas" → navigate to create form

### 8.2 Facility Create (`/fasilitas/create`)

- Enter Nama Fasilitas (required)
- Enter Tipe (required, e.g., Kesehatan, Pendidikan)
- Enter Deskripsi (optional)
- Enter Latitude (-90 to 90, required)
- Enter Longitude (-180 to 180, required)
- Pick location on Map (Leaflet map picker, syncs with lat/lng inputs)
- Upload multiple Gambar (optional)
- Click "Tambah" → create, redirect to list

### 8.3 Facility Edit (`/fasilitas/:id/edit`)

- Update all fields including coordinates
- Use map picker to adjust location
- Manage images

### 8.4 Delete Facility

- Click "Hapus" → confirm dialog → delete

---

## 9. PPID (Public Information)

### 9.1 PPID Document List (`/ppid`)

- View paginated list of public information documents
- Columns: Thumbnail, Judul, Kategori, Publikasi, Dibuat/Diperbarui, Actions
- Search by title (debounced 300ms)
- Filter by category (dropdown)
- Download document: click "Download" button → downloads PDF
- Click "Edit" → navigate to edit
- Click "Hapus" → confirm dialog → delete
- Click "Tambah Dokumen" → navigate to create form

### 9.2 PPID Document Create (`/ppid/create`)

- Enter Title (required)
- Select Category (optional)
- Upload Document file (required)
- Set Publication Date
- Click "Buat" → create

### 9.3 PPID Document Edit (`/ppid/:id/edit`)

- Update all fields
- Replace document file
- Click "Simpan" → save

### 9.4 PPID Requests (`/ppid-requests`)

- View requests from citizens for document access
- Columns: Dokumen, Nama Pemohon, Email, Tujuan, Status, Disetujui, Dibatalkan, Tanggal, Actions
- Filter by status: Semua, Menunggu, Disetujui, Dibatalkan
- Approve request: click "Setujui" → confirm → approve
- Revoke request: click "Batalkan" → confirm → revoke
- Status badges: yellow (pending), green (approved), red (revoked)

---

## 10. Organization Structure

### 10.1 View Structure (`/struktur`)

- View paginated table of organization members
- Columns: Foto, Nama, Jabatan, Actions
- Search by name or position (debounced)
- Click "Tambah Anggota" → open create modal

### 10.2 Add Member

- Fill modal form: Nama (required), Jabatan, Email, Telepon, Deskripsi
- Upload Profile Foto
- Click "Tambah" → create, close modal, refresh

### 10.3 Edit Member

- Click "Edit" on row → open edit modal with pre-filled data
- Update any fields
- Click "Simpan Perubahan" → save, close modal

### 10.4 Delete Member

- Click "Hapus" → confirm dialog → delete

---

## 11. Village Profile

### 11.1 View Profile (`/profile`)

- View village information: Nama Desa, Alamat, Telepon, Email, Website, Visi & Misi
- Click "Edit" → enter edit mode

### 11.2 Edit Profile

- Form with fields: Nama Desa (required), Alamat, Telepon, Email, Website
- Visi & Misi textarea
- Click "Simpan" → save changes
- Click "Batal" → cancel edit

---

## 12. Profile Sections

### 12.1 List Profile Sections (`/profile/sections`)

- View paginated table of profile content sections
- Columns: Nama Bagian, Konten (truncated), Status, Dibuat, Actions
- Search by section name (debounced)
- Filter by section name dropdown
- Filter by status: Aktif/Inaktif
- Click row "Edit" → navigate to edit
- Click "Hapus" → confirm dialog → delete
- Click "Tambah Informasi Detail" → navigate to create form

### 12.2 Create Profile Section (`/profile/sections/create`)

- Enter Nama Bagian (required, e.g., keuangan, lokasi-desa)
- Endpoint auto-generated (disabled field)
- Toggle Status (checkbox, active by default)
- Write konten using Rich Text Editor (required)
- Click "Tambah Informasi Detail" → create, redirect to list

### 12.3 Edit Profile Section (`/profile/sections/:id/edit`)

- Edit Nama Bagian, Status, Konten
- Endpoint is read-only
- Click "Simpan Perubahan" → save

### 12.4 Delete Profile Section

- Click "Hapus" → confirm dialog → delete

---

## 13. Infographic Dashboard

### 13.1 List Infographics (`/infographic`)

- View paginated table of Metabase dashboard embeddings
- Columns: Nama Bagian, Tipe, Component ID, Status, Dibuat, Actions
- Search by name (debounced)
- Filter by type: Dashboard, Question
- Filter by status: Aktif, Tidak Aktif
- Click "Preview" → open preview modal with embedded Metabase dashboard/question
- Click "Edit" → navigate to edit
- Click "Hapus" → confirm dialog → delete
- Click "Tambah Dashboard" → navigate to create form

### 13.2 Create Infographic (`/infographic/create`)

- Split view: Form (left) + Preview (right)
- Enter Nama Bagian (required)
- Endpoint auto-filled (must start with /)
- Select Tipe Component: Dashboard or Question (required)
- Enter Component ID (required, integer from Metabase)
- Toggle Status aktif/inaktif
- Click "Preview" → generate token, show embedded Metabase component
- Click "Tambah Dashboard" → create, redirect to list

### 13.3 Edit Infographic (`/infographic/:id/edit`)

- Edit all fields
- Preview auto-loads on component change
- Click "Simpan Perubahan" → save

### 13.4 Delete Infographic

- Click "Hapus" → confirm dialog → delete

### 13.5 Preview Modal

- Modal shows embedded Metabase dashboard or question
- Token generated per preview request
- Close by clicking X or outside modal

---

## 14. Internationalization

### 14.1 Language Support

- Supported locales: Indonesian (id), English (en)
- Language switcher in header
- Language preference saved in cookie
- All UI text uses i18n keys

---

## 15. UI Components

### 15.1 AppButton

- Variants: primary, secondary, danger
- Sizes: default, sm
- States: default, loading (with spinner), disabled
- Used for: form submissions, actions

### 15.2 AppInput

- Label display with required indicator (*)
- Placeholder text
- Error state with message display
- Types: text, email, password, etc.

### 15.3 AppModal

- Header with title
- Close button (X)
- Footer slot for actions
- Sizes: sm, md, lg
- Overlay click to close

### 15.4 AppNotification (Toast)

- Types: success, error, warning, info
- Auto-dismiss after 4 seconds
- Stack multiple notifications
- Position: top-right

### 15.5 AppPagination

- Page number buttons
- Previous/Next navigation
- Current page indicator
- Emit page-change event

### 15.6 AppTable

- Column configuration (key, label, width, sortable)
- Loading state with spinner
- Empty state with message
- Row actions via slots
- Custom cell rendering via slots

### 15.7 ConfirmDialog

- Modal with title and message
- Confirm/Cancel buttons
- Used for destructive actions

### 15.8 ImageUpload

- Preview existing image
- File input for new image
- Remove button on preview
- Emits file on change

### 15.9 FileUpload

- File input for documents
- Emits file on change

### 15.10 RichTextEditor

- Quill-based editor
- Toolbar: basic formatting, lists, links
- Used for: berita content, profile sections

### 15.11 MapPicker

- Leaflet map component
- Click to set marker
- Emits lat/lng on move
- Used for: facility location picker

### 15.12 LanguageSwitcher

- Dropdown or buttons for ID/EN
- Persists to cookie

### 15.13 LoadingSpinner

- Sizes: sm, md, lg
- Animated spinner

