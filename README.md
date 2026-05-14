# Doorman

[![Version](https://img.shields.io/github/v/tag/fishy-stick/doorman?sort=semver&filter=%21%2A-%2A&label=version)](https://github.com/fishy-stick/doorman/tags)
[![Docker Publish](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml)

Language: English | [简体中文](README.zh-CN.md)

Doorman is a lightweight DDNS server for home networks with dynamic public IPs. It is typically deployed on a public server outside the home network, receives HTTP requests from clients inside the home network, and triggers DDNS updates when the observed public IP changes.

The project includes both an admin UI and an API. In production it can run as a single Go binary, with data stored in SQLite and no external database dependency.

## Features

- Supports multiple home networks, each with its own name, token, and DDNS configuration
- Detects the public IP from the request source
- Supports `X-Forwarded-For`, `X-Real-IP`, and `RemoteAddr`
- Protects the `/knock` endpoint with a Bearer token
- Records IP history, change status, and DDNS execution results
- Currently ships with a built-in `DNSPod` provider
- Generates `curl` and `crontab` commands directly in the admin UI
- Persists data in SQLite so records survive restarts

## Requirements

- Go `1.26+`
- Node.js `22+`
- `pnpm`

If `pnpm` is not enabled on your machine yet, run:

```bash
corepack enable
```

## Quick Start

### Development Mode

In development, start the backend and frontend separately:

```bash
cp config.example.yaml config.yaml
```

```bash
go run ./cmd/doorman/
```

```bash
cd web
pnpm install
pnpm dev
```

Open `http://127.0.0.1:15173/admin/`.

Notes:

- Run `go run ./cmd/doorman/` from a directory that contains `config.yaml`
- In development, the admin UI is served by Vite while the backend handles only `/admin/api` and `/knock`
- The initial admin password is printed to the server logs on first startup

### Production Build

In production, build the frontend assets first and then embed them into the Go binary:

```bash
cd web
pnpm install
pnpm run build:embed
cd ..

go build -tags embedweb -o doorman ./cmd/doorman/
./doorman
```

Open `http://127.0.0.1:8080/admin`.

If you build without the `embedweb` tag, the binary will not include the admin UI and `/admin` will not serve the frontend.

## Configuration

Runtime configuration is provided through `config.yaml`. Home networks and DDNS rules are managed from the web UI.

Example:

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "doorman.db"
  public_url: "http://your-server:8080"
```

Field reference:

| Key | Description |
|-----|-------------|
| `server.port` | Listening port. Default: `:8080`. You can write `8080` or `:8080`; the program normalizes it automatically. |
| `server.trust_proxy` | Whether to trust proxy headers. Default: `true`. |
| `server.db` | SQLite database path. Default: `doorman.db`. |
| `server.public_url` | External service URL used to generate client commands. Default: `http://your-server:8080`. It may include a path prefix, such as `https://www.abc.com/prefix`. Query strings and fragments are not supported. |

## Typical Flow

1. Start the server and read the initial admin password from the logs.
2. Open `/admin` and sign in.
3. Create a network.
4. Copy the generated `curl` or `crontab` command.
5. Call `/knock` on the public server from the target home network on a schedule.
6. Use the admin UI to inspect the current IP, history, and DDNS status.

Doorman generates the Bearer token automatically when you create a network. If you regenerate the token later, existing client commands stop working immediately and must be updated.

## Client Request

Clients only need to send a standard HTTP request:

```bash
curl -H "Authorization: Bearer your-token" http://your-server:8080/knock
```

The admin UI generates this command from `server.public_url`. If you set `server.public_url` to `https://www.abc.com/prefix`, the generated target is `https://www.abc.com/prefix/knock`; configure your reverse proxy to forward that path to Doorman's `/knock` endpoint.

The response includes the detected IP, whether it changed, and whether DDNS was updated during that request.

## Runtime Behavior

### IP Resolution

When `server.trust_proxy=true`, Doorman resolves the client IP in this order:

1. The first valid IP in `X-Forwarded-For`
2. `X-Real-IP`
3. `RemoteAddr`

If the service is exposed directly to the public internet and is not behind a reverse proxy you control, set `trust_proxy` to `false` to avoid forged headers affecting IP detection.

### Sessions and Login

- Admin sessions are stored in memory as cookies
- Each session is valid for 24 hours
- All admin sessions are lost after a service restart
- Changing the admin password clears every active session

### DDNS Execution

- DDNS runs only when the network has DDNS enabled and the observed IP differs from the previous one
- The only built-in provider right now is `DNSPod`
- Even when DDNS is disabled, Doorman still records `/knock` history

## Deployment

Doorman supports Docker, direct binary execution, and `systemd` service management. See [deploy/README.md](deploy/README.md).

## Design Principles

- **Do not trust client-reported IPs**: determine the public IP from the request source only
- **Keep it simple**: single binary, SQLite, minimal dependencies
- **Optimize for self-hosted home use**: focus on dynamic public IP management for home networks

## License

MIT
