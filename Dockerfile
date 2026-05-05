# syntax=docker/dockerfile:1

FROM node:24-alpine AS frontend-build
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/postcube .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app
ENV PORT=8113 \
    PUBLIC_DIR=/app/public \
    DB_PATH=/app/data/postcube.db \
    ALLOWED_ORIGINS=*
COPY --from=backend-build /out/postcube /app/postcube
COPY --from=frontend-build /src/frontend/dist /app/public
EXPOSE 8113
CMD ["/app/postcube"]
