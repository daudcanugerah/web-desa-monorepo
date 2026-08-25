---
inclusion: always
---

# Git Best Practices

## Commit Message Format

Use Conventional Commits format:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer]
```

### Types
- `feat` — new feature
- `fix` — bug fix
- `refactor` — code change that neither fixes a bug nor adds a feature
- `style` — formatting, missing semicolons, etc (no logic change)
- `test` — adding or updating tests
- `chore` — build process, dependency updates, tooling
- `docs` — documentation only
- `perf` — performance improvement

### Scope (optional)
Match the feature area: `auth`, `users`, `roles`, `banners`, `berita`, `umkm`, `fasilitas`, `ppid`, `struktur`, `desa`, `router`, `store`, `api`, `ui`

### Examples
```
feat(auth): add JWT refresh token interceptor
fix(users): correct email validation on create form
refactor(api): centralize error handling in axios interceptor
chore: add vite config with vue plugin
feat(berita): implement rich text editor with TipTap
```

## Branching Strategy

- `main` — production-ready code only
- `develop` — integration branch for features
- `feat/<scope>-<short-description>` — feature branches
- `fix/<scope>-<short-description>` — bug fix branches

Branch off `develop`, merge back to `develop` via PR. Only merge `develop` → `main` for releases.

## Commit Granularity

- One logical change per commit — do not bundle unrelated changes
- Commit working code — never commit broken builds
- Each major feature (auth, user management, banner management, etc.) should be a separate commit or PR
- Infrastructure/config commits (vite, tailwind, router setup) should be separate from feature commits

## .gitignore

Always exclude:
```
node_modules/
dist/
.env
.env.local
.env.*.local
*.log
.DS_Store
```

## Before Committing

1. Ensure the app builds without errors (`npm run build`)
2. Ensure no console errors in development
3. Stage only relevant files — avoid committing unrelated changes with `git add -p` if needed

## Pull Requests

- Keep PRs focused — one feature or fix per PR
- Write a clear PR description referencing the requirement it satisfies
- Squash trivial fixup commits before merging
