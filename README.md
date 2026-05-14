# Doorman

[![Version](https://img.shields.io/github/v/tag/fishy-stick/doorman?sort=semver&filter=%21%2A-%2A&label=version)](https://github.com/fishy-stick/doorman/tags)
[![Docker Publish](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml)

Language: English | [简体中文](README.zh-CN.md)

Doorman is a lightweight DDNS server for home networks with dynamic public IPs. It runs on a public server outside the home network, receives HTTP requests from clients inside the home network, and triggers DDNS updates when the observed public IP changes.

The project includes a web admin UI and an API. In production it can run as a single Go binary, with data stored in SQLite and no external database dependency.

## Use Cases

- Your home broadband has a dynamic public IP and needs automatic DNS updates
- You want the DDNS server to run on a VPS, cloud instance, or another long-running public node
- You want clients to make simple HTTP requests without exposing extra services inside the home network
- You need to manage multiple home networks or broadband lines
- You use, or can accept using, the built-in `DNSPod` DDNS provider

## Highlights

- Supports multiple home networks, each with its own name, token, and DDNS configuration
- Detects the public IP from the request source instead of trusting client-reported IPs
- Supports `X-Forwarded-For`, `X-Real-IP`, and `RemoteAddr`
- Protects the `/knock` endpoint with a Bearer token
- Records IP history, change status, and DDNS execution results
- Currently ships with a built-in `DNSPod` provider
- Generates `curl` and `crontab` commands directly in the admin UI
- Persists data in SQLite so records survive restarts

## Deployment Options

Doorman needs to be deployed on a public server outside the home network. Docker deployment is recommended; direct binary execution and `systemd` service management are also supported.

For complete deployment instructions, see [deploy/README.md](deploy/README.md).

## 5-Minute Quick Deploy

The shortest trial path is to run Doorman with Docker on a public server.

### 1. Build the Image

Run this from the repository root:

```bash
docker build -t doorman .
```

### 2. Start the Service

For a quick trial, you can use the configuration built into the image:

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v doorman-data:/app/data \
  doorman
```

For a real deployment, mount your own configuration file and set `server.public_url` to the real external URL:

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/app/data/doorman.db"
  public_url: "https://your-domain.example"
```

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  -v doorman-data:/app/data \
  doorman
```

`server.public_url` is used to generate client `curl` and `crontab` commands. The default value, `http://your-server:8080`, is only suitable for a trial run and should be changed before real use.

### 3. Get the Initial Admin Password

On first startup, Doorman generates an admin password and prints it to the logs:

```bash
docker logs doorman
```

Then open:

```text
http://<your-host>:8080/admin
```

### 4. Create a Network

After signing in to the admin UI:

1. Create a network.
2. Enable DDNS if needed and fill in the DNSPod configuration.
3. Copy the generated `curl` or `crontab` command.
4. Run that command on a schedule from inside the target home network.

Doorman generates the Bearer token automatically when you create a network. If you regenerate the token later, existing client commands stop working immediately and must be updated.

## Client Request

Clients only need to send a standard HTTP request from inside the target home network:

```bash
curl -H "Authorization: Bearer your-token" https://your-domain.example/knock
```

The admin UI generates this command from `server.public_url`. If you set `server.public_url` to `https://www.abc.com/prefix`, the generated target is `https://www.abc.com/prefix/knock`; configure your reverse proxy to forward that path to Doorman's `/knock` endpoint.

The response includes the detected IP, whether it changed, and whether DDNS was updated during that request.

## Core Configuration

Runtime configuration is provided through `config.yaml` in the current working directory. Home networks and DDNS rules are managed from the web UI.

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

If Doorman is exposed directly to the public internet and is not behind a reverse proxy you control, set `trust_proxy` to `false` to avoid forged headers affecting IP detection.

## Runtime Behavior

### IP Resolution

When `server.trust_proxy=true`, Doorman resolves the client IP in this order:

1. The first valid IP in `X-Forwarded-For`
2. `X-Real-IP`
3. `RemoteAddr`

### Sessions and Login

- Admin sessions are stored in memory as cookies
- Each session is valid for 24 hours
- All admin sessions are lost after a service restart
- Changing the admin password clears every active session

### DDNS Execution

- DDNS runs only when the network has DDNS enabled and the observed IP differs from the previous one
- The only built-in provider right now is `DNSPod`
- Even when DDNS is disabled, Doorman still records `/knock` history

## Local Development and Build

### Requirements

- Go `1.26+`
- Node.js `22+`
- `pnpm`

If `pnpm` is not enabled on your machine yet, run:

```bash
corepack enable
```

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

## Design Principles

- **Do not trust client-reported IPs**: determine the public IP from the request source only
- **Keep it simple**: single binary, SQLite, minimal dependencies
- **Optimize for self-hosted home use**: focus on dynamic public IP management for home networks

## License

MIT
