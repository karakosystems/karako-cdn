<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/assets/logo-full-white.png">
    <img src=".github/assets/logo-full-navy.png" alt="Karako" width="300">
  </picture>
</p>

# karako-cdn

A static asset CDN as a Docker image. It loads a folder into memory at
startup and serves it with immutable caching, ETag/304, gzip, Range and
open CORS. Scratch image, zero dependencies. Inspired by
[coolify-cdn](https://github.com/coollabsio/coolify-cdn).

## Usage

```dockerfile
FROM ghcr.io/karakosystems/karako-cdn:1
ENV BASE_FQDN=example.com
COPY --chown=65534:65534 assets/ /assets/
```

```bash
docker build -t my-cdn . && docker run -p 8080:8080 my-cdn
```

**The path under `assets/` is the URL**: `assets/brand/logo.png` is
served at `/brand/logo.png`. Dotfiles are never served. Files are cached
for a year, so change the URL (`logo.png?v=2`) when one changes.

## Open Graph images

The image ships `oggen`, which renders social cards from an
`og.config.json` sitting next to `assets/`:

```dockerfile
COPY og.config.json /og.config.json
RUN ["/oggen", "-root", "/"]
```

Each project in the config produces `assets/<name>/og-image.png`, or one
per language when it declares `locales`. See [`example/`](example).

## Endpoints

- `GET /<path>`: asset (Range, gzip)
- `GET /discover.json`: every resource, with absolute URLs
- `GET /health` · `GET /metrics`: probe and Prometheus counters
- unknown path: `404` when it has an extension, otherwise `302` to
  `https://{BASE_FQDN}`

## Configuration

| Variable     | Default                      | Purpose                          |
| ------------ | ---------------------------- | -------------------------------- |
| `BASE_FQDN`  | required                     | redirect target                  |
| `CDN_FQDN`   | `cdn.{BASE_FQDN}`            | absolute URLs in `discover.json` |
| `ASSETS_DIR` | `assets` (Docker: `/assets`) | folder served                    |
| `PORT`       | `80` (Docker: `8080`)        | listen port                      |

The image runs as `nobody` (65534) on 8080, drains on SIGTERM, and is
tagged `1.0.0`, `1.0`, `1` and `latest` for amd64 and arm64.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md),
[GOVERNANCE.md](.github/GOVERNANCE.md) and the
[Code of Conduct](.github/CODE_OF_CONDUCT.md). Vulnerabilities go to
[SECURITY.md](.github/SECURITY.md), never to a public issue.

## License

[MIT](LICENSE)
