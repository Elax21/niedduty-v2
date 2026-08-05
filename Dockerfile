# Ein Artefakt: Frontend wird gebaut und ins Go-Binary eingebettet.

# 1) Frontend bauen → internal/web/dist
FROM docker.io/library/node:22-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci
COPY frontend ./frontend
RUN cd frontend && npm run build

# 2) Server bauen (mit eingebettetem Frontend)
FROM docker.io/library/golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY --from=frontend /src/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/niedduty ./cmd/server

# 3) Laufzeit — Zertifikate für fussball.de und die Push-Dienste nötig
FROM docker.io/library/alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 niedduty
ENV TZ=Europe/Berlin PRODUCTION=true LISTEN_ADDR=:8080
USER niedduty
COPY --from=backend /out/niedduty /usr/local/bin/niedduty
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/niedduty"]
