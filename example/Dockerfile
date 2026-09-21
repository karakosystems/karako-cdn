ARG BASE_IMAGE=ghcr.io/karakosystems/karako-cdn:latest
FROM ${BASE_IMAGE}

ENV BASE_FQDN=example.com

COPY --chown=65534:65534 public/ /public/
# Optional: render public/<name>/og-image.png from og.config.json at build time
COPY og.config.json /og.config.json
RUN ["/oggen", "-root", "/"]
