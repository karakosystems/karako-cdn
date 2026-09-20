<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/assets/logo-full-white.png">
    <img src=".github/assets/logo-full-navy.png" alt="Karako" width="300">
  </picture>
</p>

# karako-cdn

A static asset CDN packaged as a Docker image: a Go server that loads a
folder of files into memory at startup and serves it with immutable
caching, ETag/304, gzip, Range and open CORS. Scratch image, zero
dependencies. Inspired by
[coolify-cdn](https://github.com/coollabsio/coolify-cdn).

## Usage

Give your project an `assets/` folder and this Dockerfile:

```dockerfile
FROM ghcr.io/karakosystems/karako-cdn:0.1
ENV BASE_FQDN=example.com
COPY --chown=65534:65534 assets/ /assets/
```

```bash
docker build -t my-cdn .
docker run -p 8080:8080 my-cdn
```

**The path under `assets/` is the URL**: `assets/brand/logo.png` is
served at `/brand/logo.png`. Dotfiles are never served. Assets are
cached for a year, so change the URL (`logo.png?v=2`) when a file
changes. Mounting a folder (`-v ./assets:/assets:ro`) also works, handy
for local development.

### Open Graph images

The image ships `oggen`, which renders 1200x630 social cards from an
`og.config.json` placed next to `assets/`. Add two lines to render them
during the build:

```dockerfile
COPY og.config.json /og.config.json
RUN ["/oggen", "-root", "/"]
```

Each project in the config produces `assets/<name>/og-image.png` (plus
`og-image-<lang>.png` when it declares `locales`). See
[`example/`](example) for a complete project and the config format.
Logo paths in the config are relative to the folder holding it (for
example `assets/brand/logo-full-white.png`).

## Endpoints

- `GET /<path>`: asset (`max-age=31536000, immutable`, Range, gzip)
- `GET /discover.json`: compact list of every resource (absolute URLs)
- `GET /health` · `GET /metrics`: probe and Prometheus counters
- unknown path: cacheable `404` (with extension), `302` to
  `https://{BASE_FQDN}` (without)

## Configuration

| Variable     | Default                      | Purpose                          |
| ------------ | ---------------------------- | -------------------------------- |
| `BASE_FQDN`  | required                     | redirect target                  |
| `CDN_FQDN`   | `cdn.{BASE_FQDN}`            | absolute URLs in `discover.json` |
| `ASSETS_DIR` | `assets` (Docker: `/assets`) | folder served                    |
| `PORT`       | `80` (Docker: `8080`)        | listen port                      |

The image runs as `nobody` (65534) on 8080 and drains connections on
SIGTERM. Tags follow semver (`0.1.0`, `0.1`, `latest`) for `linux/amd64`
and `linux/arm64`.

## Development

```bash
go test ./...                              # unit tests
go -C tools/oggen test ./...               # oggen tests (separate module)
BASE_FQDN=example.com PORT=8080 ASSETS_DIR=example/assets go run ./cmd/karako-cdn
docker compose up --build                  # serves example/assets on :8080
./test.sh 8080                             # integration suite
lefthook install                           # git hooks
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org):
release-please turns them into versions, the changelog and a new image
on GHCR. Contributions are welcome through pull requests.

## License

[MIT](LICENSE)
