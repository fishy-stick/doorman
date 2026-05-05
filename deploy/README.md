# Deploy

Language: English | [简体中文](README.zh-CN.md)

This guide covers the three supported deployment modes for Doorman:

- Docker
- Direct binary execution
- `systemd` service management

Doorman is shipped as a Go binary with embedded frontend assets. At runtime it depends on `config.yaml` and a SQLite database file. It is best deployed on a publicly reachable server outside the home network, such as a VPS, cloud instance, or another long-running public node.

## Docker Deployment

### Build the Image

Run this from the repository root:

```bash
docker build -t doorman .
```

The image build does the following:

- Builds the frontend assets from `web/`
- Embeds the frontend output into the Go binary
- Writes a default `/app/config.yaml` into the runtime image

### Default Runtime

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v doorman-data:/app/data \
  doorman
```

The default in-container configuration is equivalent to:

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/app/data/doorman.db"
```

Default runtime properties:

- The service listens on container port `8080`
- The database is stored at `/app/data/doorman.db`
- The container runs as `nobody:nogroup`
- The image declares `VOLUME /app/data`

As long as `/app/data` is persisted, the database survives container recreation.

### Custom Configuration

If you need to change the listening port, database path, or `trust_proxy`, mount your own configuration file:

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  -v doorman-data:/app/data \
  doorman
```

If you move `server.db` to another path, make sure the matching directory is mounted as well.

If you use a bind mount instead of a Docker volume, ensure the mounted directory is writable by `nobody:nogroup`, otherwise SQLite cannot create or update the database.

### First Startup and Checks

On first startup, the service generates an admin password and prints it to the logs:

```bash
docker logs doorman
```

Then open:

```text
http://<your-host>:8080/admin
```

Recommended initial checks:

1. Confirm `/admin` opens successfully
2. Sign in with the password from the logs
3. Create a network and trigger `/knock` once
4. Verify `/app/data/doorman.db` exists and history records are visible

## Direct Binary Execution

### Build

Build the frontend first, then compile the embedded binary:

```bash
cd web
pnpm install
pnpm run build:embed
cd ..

go build -tags embedweb -o doorman ./cmd/doorman/
```

### Prepare Configuration

Create `config.yaml` in the binary working directory:

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/var/lib/doorman/doorman.db"
```

The program always reads `config.yaml` from the current working directory. If you launch it with:

```bash
./doorman
```

then your current shell directory must contain `config.yaml`.

If `server.db` uses a relative path such as `doorman.db`, that path is also resolved relative to the current working directory. Absolute paths are safer for production.

### Start

```bash
./doorman
```

After first startup, read the initial admin password from standard output logs.

## Running Under systemd

Here is a `systemd` unit example you can adapt directly:

```ini
[Unit]
Description=Doorman DDNS Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=doorman
Group=doorman
WorkingDirectory=/opt/doorman
ExecStart=/opt/doorman/doorman
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Suggested directory layout:

- `/opt/doorman/doorman`
- `/opt/doorman/config.yaml`
- `/var/lib/doorman/doorman.db`

Recommended setup:

- Create a dedicated service user such as `doorman`
- Use an absolute database path
- Set `WorkingDirectory` to the directory that contains `config.yaml`
- Make sure the database directory is writable by the service user

Common commands:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now doorman
sudo systemctl status doorman
sudo journalctl -u doorman -f
```

On first startup, read the initial admin password from `journalctl`.

## Reverse Proxy and trust_proxy

If Doorman is behind Nginx, Caddy, or another reverse proxy that you control, you can keep:

```yaml
server:
  trust_proxy: true
```

In that mode, Doorman trusts these sources in order:

1. `X-Forwarded-For`
2. `X-Real-IP`
3. `RemoteAddr`

If the service is exposed directly to the public internet, or the upstream proxy chain is not fully trusted, use:

```yaml
server:
  trust_proxy: false
```

Otherwise, forged headers may affect public IP detection.

## Upgrade and Change Notes

- When upgrading Docker deployments, keep the database file in the existing volume or bind mount
- After regenerating a network token, all existing client commands stop working immediately
- Admin sessions are stored in memory, so a restart requires logging in again
- Before changing `trust_proxy`, confirm the real network topology and proxy chain
