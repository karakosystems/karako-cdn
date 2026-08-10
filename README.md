<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/karako/logos/logo-full-white.png">
    <img src="assets/karako/logos/logo-full-navy.png" alt="Karako" width="300">
  </picture>
</p>

# karako-cdn

Karako's static asset CDN: a Go server that embeds `assets/` into the
binary. Scratch Docker image (~10 MB), zero dependencies.

**The path under `assets/` is the URL** — drop a file (lowercase
kebab-case), rebuild, it is served with immutable caching, ETag/304,
gzip and open CORS. Dotfiles are never served.

## Endpoints

- `GET /<path>` — asset (`max-age=31536000, immutable`, Range, gzip)
- `GET /discover.json` — compact list of every resource (absolute URLs)
- `GET /health` · `GET /metrics` — probe and Prometheus counters
- unknown path: cacheable `404` (with extension), `302` to
  `https://{BASE_FQDN}` (without)

## Usage

```bash
go test ./...                        # unit + embed tests
PORT=8080 go run ./cmd/karako-cdn    # run locally
docker compose up --build            # dev container on :8080
./test.sh 8080                       # integration suite
go -C tools/oggen run .              # regenerate OG images
```

The image runs as `nobody` on 8080 and drains connections on SIGTERM.
Projects with `locales` in `og.config.json` get one OG image per
language.

## Configuration

| Variable    | Default            | Purpose                          |
| ----------- | ------------------ | -------------------------------- |
| `BASE_FQDN` | `karakosystems.com`| redirect target                  |
| `CDN_FQDN`  | `cdn.{BASE_FQDN}`  | absolute URLs in `discover.json` |
| `PORT`      | `80` (Docker: `8080`) | listen port                   |
