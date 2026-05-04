# syntax=docker/dockerfile:1

FROM node:22-bookworm AS web-builder
WORKDIR /src/web

COPY web/ ./

RUN corepack enable \
    && pnpm install --frozen-lockfile \
    && pnpm run build:embed

FROM golang:1.26-bookworm AS go-builder
WORKDIR /src

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY --from=web-builder /src/internal/webui/dist ./internal/webui/dist

RUN go build -tags embedweb -o /out/doorman ./cmd/doorman

FROM debian:bookworm-slim AS runtime
WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /out/doorman /usr/local/bin/doorman

RUN mkdir -p /app/data \
    && printf 'server:\n  port: 8080\n  trust_proxy: true\n  db: "/app/data/doorman.db"\n' > /app/config.yaml \
    && chown -R nobody:nogroup /app

USER nobody:nogroup

EXPOSE 8080
VOLUME ["/app/data"]

CMD ["doorman"]
