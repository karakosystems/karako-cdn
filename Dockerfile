FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY tools/oggen/go.mod tools/oggen/go.sum ./tools/oggen/
RUN go -C tools/oggen mod download
COPY tools/oggen/ ./tools/oggen/

ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go build -o karako-cdn ./cmd/karako-cdn
RUN go build -o healthcheck ./cmd/healthcheck
RUN go -C tools/oggen build -o /app/oggen .

FROM scratch

LABEL org.opencontainers.image.title="karako-cdn" \
      org.opencontainers.image.description="Static asset CDN as a Docker image. Drop a folder in, get immutable caching, ETag/304, gzip, Range, open CORS, /discover.json, Prometheus metrics and build-time Open Graph images. Scratch image, zero dependencies." \
      org.opencontainers.image.url="https://github.com/karakosystems/karako-cdn" \
      org.opencontainers.image.source="https://github.com/karakosystems/karako-cdn" \
      org.opencontainers.image.documentation="https://github.com/karakosystems/karako-cdn#readme" \
      org.opencontainers.image.vendor="Karako Systems" \
      org.opencontainers.image.authors="Louis Midson Lajeanty (@midsonlajeanty)" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.base.name="scratch"

COPY --from=builder /app/karako-cdn /karako-cdn
COPY --from=builder /app/healthcheck /healthcheck
COPY --from=builder /app/oggen /oggen

ENV PUBLIC_DIR=/public
ENV PORT=8080

USER 65534:65534

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

CMD ["/karako-cdn"]
