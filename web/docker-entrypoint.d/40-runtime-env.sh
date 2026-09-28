#!/bin/sh
# Generate /usr/share/nginx/html/env.js from the container environment at
# startup. nginx:alpine runs every *.sh in docker-entrypoint.d/ before nginx.
#
# Defaults come from the Dockerfile ENV (the build-time ARG values). A value
# passed at run time overrides them, so one built image can be re-pointed at a
# different API/Metabase without a rebuild:
#   docker run -e VITE_API_BASE_URL=... -e VITE_METABASE_URL=... desa-web:tag
set -eu

: "${VITE_API_BASE_URL:=https://api.desapalasari.my.id/api/v1}"
: "${VITE_METABASE_URL:=https://bi-embed.desapalasari.my.id}"
: "${VITE_FEATURE_GALLERY_PUBLIC:=true}"
: "${VITE_FEATURE_BERITA_SEARCH:=false}"
: "${VITE_FEATURE_INFOGRAPHIC_SORT:=false}"

# Escape backslashes and double quotes for JSON string safety.
json_escape() { printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'; }

cat > /usr/share/nginx/html/env.js <<EOF
window.__ENV__ = {
  "VITE_API_BASE_URL": "$(json_escape "$VITE_API_BASE_URL")",
  "VITE_METABASE_URL": "$(json_escape "$VITE_METABASE_URL")",
  "VITE_FEATURE_GALLERY_PUBLIC": "$(json_escape "$VITE_FEATURE_GALLERY_PUBLIC")",
  "VITE_FEATURE_BERITA_SEARCH": "$(json_escape "$VITE_FEATURE_BERITA_SEARCH")",
  "VITE_FEATURE_INFOGRAPHIC_SORT": "$(json_escape "$VITE_FEATURE_INFOGRAPHIC_SORT")"
};
EOF

echo "web: generated env.js (api=${VITE_API_BASE_URL})"