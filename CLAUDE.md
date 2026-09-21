# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## Overview

karako-cdn is a static CDN shipped as a Docker image, inspired by
coollabsio/coolify-cdn. A small Go server loads `PUBLIC_DIR` into memory at
startup and serves it with HTTP caching, ETag/304, gzip, Range and open CORS,
plus a generated `/discover.json`. Scratch image (amd64/arm64), zero external
dependencies in the server module.

The image is public on `ghcr.io/karakosystems/karako-cdn` and mirrored to Docker
Hub. Downstream projects reuse it with `FROM` + `COPY --chown=65534:65534
public/ /public/`, and can render their Open Graph images at build time with
`RUN ["/oggen", "-root", "/"]`. Karako's own files live in the private repo
`karakosystems/cdn.karakosystems.com`; `example/` here is the demo project used
by tests and CI.

## Conventions

Everything here is English, commit messages included: the repo is public and its
commits feed the CHANGELOG and release notes. Conventional commits are enforced
by the `commit-msg` hook. Comments stay rare, the code stays self-documenting.
The private assets repo keeps French commit messages.

File names are lowercase kebab-case with lowercase extensions. Logos follow
`logo-full-<variant>.png`, `logo-icon-<variant>.png` and `logo-text-<variant>.png`.

## Commands

```bash
go test ./...                                           # unit tests
go test ./internal/server/ -run TestETagReturns304 -v   # single test
go -C tools/oggen test ./...                            # oggen, separate module
BASE_FQDN=example.com PORT=8080 PUBLIC_DIR=example/public go run ./cmd/karako-cdn
./test.sh [port] [host]                                 # integration suite
docker build -t karako-cdn:local .
docker build --build-arg BASE_IMAGE=karako-cdn:local -t example example
lefthook install                                        # git hooks
```

## Architecture

**URL rule: the path under `PUBLIC_DIR` IS the URL.** `public/karako/x.png` is
served at `/karako/x.png`. Every non-dotfile is served, whatever its extension.

`internal/server`:

- `files.go` walks the FS into memory maps: content plus SHA-256 ETags truncated
  to 128 bits, dotfiles skipped.
- `discover.go` builds `/discover.json` at startup (compact JSON, absolute URLs
  from `CDN_FQDN`, etag values without quotes) and registers it as a regular
  resource, so it gets ETag/304/Cache-Control for free. Startup fails if a file
  claims that path.
- `compress.go` precomputes gzip variants, with their own ETag, for text-like
  types of at least 256 bytes that shrink by 10% or more.
- `metrics.go` holds atomic counters served at `/metrics` in Prometheus text
  format. Startup fails if a file claims that path.
- `server.go` defines `Config{BaseFQDN, CDNFQDN, Addr}` and `New(cfg, fs.FS)`,
  where the FS is rooted at the served folder (`os.DirFS(PUBLIC_DIR)`).
- `handler.go` is the `http.Handler`: CORS on every response, then OPTIONS 204,
  `/` redirect, `/health`, `/metrics`, `/favicon.ico` 404, known file through
  `http.ServeContent`, and finally the miss (cacheable 404 when the path has an
  extension, 302 redirect otherwise).

Caching: files are immutable per deployment, so they ship
`public, max-age=31536000, immutable`. Scripts and configuration
(`revalidatedExtensions` in `handler.go`) and `/discover.json` keep
`must-revalidate, max-age=600`; misses get `public, max-age=300`.

`cmd/karako-cdn` is the entrypoint (read-header and idle timeouts, graceful
drain on SIGTERM). `cmd/healthcheck` is a separate tiny binary used as the
Docker `HEALTHCHECK`, because the scratch image has no shell or curl.

`tools/oggen` is a separate Go module, built into the image as `/oggen`. It
renders `<root>/public/<name>/og-image.png` from `<root>/og.config.json` using
`golang.org/x/image` and the embedded Go fonts, which keeps the server module
dependency-free. A project declares either top-level `type`/`tagline`/
`description` for a single image, or `locales` + `defaultLocale` for one
`og-image-<lang>.png` per language plus `og-image.png` for the default locale.
Mixing both fails validation.

Tests use `fstest.MapFS`, except `internal/server/example_files_test.go`, which
loads the real `example/public`.

## Configuration

| Variable     | Default                      | Notes                              |
| ------------ | ---------------------------- | ---------------------------------- |
| `BASE_FQDN`  | required                     | redirect target, startup fails without it |
| `CDN_FQDN`   | `cdn.{BASE_FQDN}`            | absolute URLs in `discover.json`   |
| `PUBLIC_DIR` | `public` (image: `/public`)  | folder served                      |
| `PORT`       | `80` (image: `8080`)         | the image runs as user 65534, so map `-p <host>:8080` |

## Gotchas

- Files are read once at startup: a change needs a restart or a new image. They
  are cached for a year, so consumers bust the cache through the URL (`?v=2`).
- `og.config.json` sits next to `public/`, so it is never served. OG rendering
  happens at image build time, never at startup, and downstream `COPY` must use
  `--chown=65534:65534` because `/oggen` runs as nobody.
- `contentTypeFor` in `handler.go` resolves from its explicit table first, then
  `mime.TypeByExtension`. The scratch image has no `/etc/mime.types`, so only
  Go's builtin list applies at runtime: add uncommon types to the table.
- `gzipMinSize` (256 B) exists because below that the gzip header eats the
  savings; variants are also dropped unless they save 10% or more.
- No outbound request is made, so the image ships without CA certificates. Copy
  `/etc/ssl/certs/ca-certificates.crt` into the scratch stage if that changes.
- The Dockerfile copies an explicit file list (`go.mod`, `cmd/`, `internal/`,
  `tools/oggen/`): a new top-level directory needed by the build must be added
  there too.
- `.gitignore` anchors `/karako-cdn` and `/healthcheck` to the repo root on
  purpose: unanchored patterns would shadow the `cmd/` source directories.
- Root `go test ./...` does not cover `tools/oggen`; run its suite explicitly.
- Releases: release-please (config in `release/`) opens the release PR; merging
  it calls the reusable `release.yml`, which builds once and pushes to GHCR, and
  to Docker Hub when `DOCKERHUB_REPO` is set. `release.yml` also accepts a manual
  dispatch with a version, to republish without cutting a release.
