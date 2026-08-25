# Desa Admin Dashboard - Specification

## Overview

Admin dashboard for Village Information Management System (Sistem Informasi Desa). Built with Vue.js 3, it enables village administrators to manage public content, users, organization structure, and PPID (public information) documents.

## Documents

| Document | Description |
|----------|-------------|
| [FEATURES.md](./FEATURES.md) | Feature descriptions by domain |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Technical architecture and stack |
| [ROUTES.md](./ROUTES.md) | Route definitions and access control |

## Quick Reference

| Domain | Routes | Service |
|--------|--------|---------|
| Auth | `/login`, `/forgot-password`, `/reset-password` | `auth.service.js` |
| Dashboard | `/` | - |
| Users | `/users`, `/users/create`, `/users/:id/edit`, `/me` | `user.service.js` |
| Roles | `/roles` | `role.service.js` |
| Banners | `/banners`, `/banners/create`, `/banners/:id/edit` | `banner.service.js` |
| Berita | `/berita`, `/berita/create`, `/berita/:id/edit` | `news.service.js` |
| UMKM | `/umkm`, `/umkm/create`, `/umkm/:id/edit` | `umkm.service.js` |
| Fasilitas | `/fasilitas`, `/fasilitas/create`, `/fasilitas/:id/edit` | `facility.service.js` |
| PPID | `/ppid`, `/ppid/create`, `/ppid/:id/edit`, `/ppid-requests` | `ppid.service.js` |
| Struktur | `/struktur` | `structure.service.js` |
| Profile | `/profile`, `/profile/sections`, `/profile/sections/create`, `/profile/sections/:id/edit` | `profile.service.js`, `profile-section.service.js` |
| Infographic | `/infographic`, `/infographic/create`, `/infographic/:id/edit` | `infographic.service.js` |

## API Reference

API specification is defined in `swagger.yaml` (OpenAPI 3.0.3).

Base URL: `VITE_API_BASE_URL` environment variable (default: `http://localhost:8080`)