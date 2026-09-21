<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/assets/logo-full-white.png">
    <img src=".github/assets/logo-full-navy.png" alt="Karako" width="300">
  </picture>
</p>

# karako-cdn

A static asset CDN as a Docker image. It loads a folder into memory at
startup and serves it with caching, ETag/304, gzip, Range and open CORS.
Scratch image, zero dependencies. Inspired by
[coolify-cdn](https://github.com/coollabsio/coolify-cdn).

```dockerfile
FROM ghcr.io/karakosystems/karako-cdn:1
ENV BASE_FQDN=example.com
COPY --chown=65534:65534 public/ /public/
```

**The path under `public/` is the URL**: `public/brand/logo.png` is
served at `/brand/logo.png`. Any file type is served, dotfiles never
are. Files are cached for a year, so change the URL (`logo.png?v=2`)
when one changes; scripts and configuration (`.sh`, `.json`, `.yaml`,
`.toml`, ...) instead revalidate every 10 minutes, so
`curl -fsSL https://cdn.example.com/install.sh | sh` stays current.

Besides your files, the server answers `/discover.json` (every resource,
with absolute URLs), `/health` and `/metrics` (Prometheus). Unknown
paths get a `404` when they have an extension, otherwise a `302` to
`BASE_FQDN`.

## Configuration

| Variable     | Default                      | Purpose                          |
| ------------ | ---------------------------- | -------------------------------- |
| `BASE_FQDN`  | required                     | redirect target                  |
| `CDN_FQDN`   | `cdn.{BASE_FQDN}`            | absolute URLs in `discover.json` |
| `PUBLIC_DIR` | `public` (Docker: `/public`) | folder served                    |
| `PORT`       | `80` (Docker: `8080`)        | listen port                      |

The image runs as `nobody` (65534) on 8080, drains on SIGTERM, and is
tagged `1.0.0`, `1.0`, `1`, `latest` and `sha-<commit>` for amd64 and
arm64.

## Open Graph images

The image ships `oggen`, which renders social cards from an
`og.config.json` sitting next to `public/`:

```dockerfile
COPY og.config.json /og.config.json
RUN ["/oggen", "-root", "/"]
```

Each project in the config produces `public/<name>/og-image.png`, or one
per language when it declares `locales`. See [`example/`](example) for a
complete project.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md),
[GOVERNANCE.md](.github/GOVERNANCE.md) and the
[Code of Conduct](.github/CODE_OF_CONDUCT.md). Vulnerabilities go to
[SECURITY.md](.github/SECURITY.md), never to a public issue.

## License

[MIT](LICENSE)
