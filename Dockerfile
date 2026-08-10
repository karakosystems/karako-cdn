# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY embed.go ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY assets/ ./assets/

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o karako-cdn ./cmd/karako-cdn
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o healthcheck ./cmd/healthcheck

FROM scratch

COPY --from=builder /app/karako-cdn /karako-cdn
COPY --from=builder /app/healthcheck /healthcheck

ARG BASE_FQDN=karakosystems.com
ENV BASE_FQDN=$BASE_FQDN
ENV PORT=8080

USER 65534:65534

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

CMD ["/karako-cdn"]
