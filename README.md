# karako-cdn

Karako's static asset CDN: a small Go server that embeds `assets/` into
the binary and serves it with aggressive HTTP caching, gzip, open CORS
and redirects. Scratch Docker image (~10 MB), AMD64/ARM64, zero
dependencies.

**URL rule: the path under `assets/` is the URL.**
`assets/images/karako/logos/logo-icon-black.png` →
`/images/karako/logos/logo-icon-black.png`.

To publish a file, drop it under `assets/` (lowercase kebab-case) and
rebuild — content is embedded at compile time. Dotfiles are never
served; embedded directories must not be empty (keep a `.gitkeep`).

## Endpoints

- `GET /<path under assets/>` — ETag/304, Range, gzip when accepted,
  `Cache-Control: public, max-age=31536000, immutable`
- `GET /discover.json` — compact JSON listing every resource (absolute
  URL, content type, size, etag), `max-age=600`
- `GET /health` — `200 healthy`
- `GET /metrics` — Prometheus counters
- unknown path with extension — cacheable `404`; without — `302` to
  `https://{BASE_FQDN}`
- `OPTIONS *` — `204`, `Access-Control-Allow-Origin: *`

## Development

```bash
go build ./...
go test ./...
PORT=8080 go run ./cmd/karako-cdn
go -C tools/oggen run .    # regenerate OG images from og.config.json
```

## Docker

```bash
docker compose up --build    # dev on :8080
docker compose watch         # rebuild on assets/ or code change
./test.sh 8080

docker buildx build --platform linux/amd64,linux/arm64 -t karako-cdn .
```

The image runs as `nobody` (65534) and listens on 8080. Drains
connections on SIGTERM.

## Configuration

- `BASE_FQDN` (default `karakosystems.com`) — redirect target.
- `CDN_FQDN` (default `cdn.{BASE_FQDN}`) — host used for the absolute
  URLs in `discover.json`.
- `PORT` (default `80`; the Docker image sets `8080`).
