# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM node:22-alpine AS build
WORKDIR /app

# Install deps first for better layer caching
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# API origin (host only — the admin client appends /api/v1 itself).
ARG VITE_API_BASE_URL=https://api.desapalasari.my.id
ARG VITE_METABASE_URL=https://bi-embed.desapalasari.my.id

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_METABASE_URL=$VITE_METABASE_URL

RUN npm run build

# ---- Runtime stage ----
FROM nginx:1.27-alpine AS runtime

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
