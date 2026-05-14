# 部署

语言： [English](README.md) | 简体中文

本文档覆盖 Doorman 的生产部署方式：

- GHCR 容器镜像，推荐优先使用
- GitHub Release 二进制归档，不使用容器时推荐优先使用
- `systemd` 托管
- 源码构建，用于开发、调试或兜底场景

Doorman 推荐部署在家庭网络之外、具备公网可达性的服务器上，例如 VPS、云主机或其他长期在线的公网节点。生产形态是一个嵌入前端资源的 Go 二进制，运行时依赖 `config.yaml` 和 SQLite 数据文件。

## 部署前准备

部署前先确认：

- 服务器可以被目标家庭网络访问
- 已确定外部访问地址，并准备写入 `server.public_url`
- 已确定是否位于可信反向代理后面
- SQLite 数据文件所在目录会被持久化
- 防火墙或安全组已放行对外服务端口，默认是 `8080`

`server.public_url` 会影响后台生成的 `curl` 和 `crontab` 命令。正式部署时不要保留默认的 `http://your-server:8080`。

## GHCR 容器部署

优先使用 GitHub Container Registry 上的预构建镜像：

```bash
docker pull ghcr.io/fishy-stick/doorman:<version>
```

将 `<version>` 替换成已发布的版本标签，例如 `v0.2-alpha`。稳定版也会发布 `latest` 和语义化别名；预发布版本应使用精确标签。

### 快速试跑

如果只是先确认服务可以启动，可以使用镜像内置配置：

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v doorman-data:/app/data \
  ghcr.io/fishy-stick/doorman:<version>
```

镜像内默认配置等价于：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/app/data/doorman.db"
  public_url: "http://your-server:8080"
```

默认约束：

- 服务监听容器内 `8080`
- 数据库存放在 `/app/data/doorman.db`
- 运行用户是 `nobody:nogroup`
- 镜像声明了 `VOLUME /app/data`

只要 `/app/data` 做了持久化挂载，容器重建后数据库仍会保留。

### 正式部署配置

正式部署建议准备自己的 `config.yaml`：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/app/data/doorman.db"
  public_url: "https://your-domain.example"
```

然后挂载配置文件和数据卷：

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  -v doorman-data:/app/data \
  ghcr.io/fishy-stick/doorman:<version>
```

如果你把 `server.db` 改到别的位置，记得同步挂载对应目录。

如果使用 bind mount 而不是 Docker volume，确保挂载目录对容器内的 `nobody:nogroup` 可读写，否则 SQLite 无法创建或更新数据库。

### 首次启动与检查

首次启动时，服务会自动生成管理员密码并输出到日志：

```bash
docker logs doorman
```

拿到密码后访问：

```text
http://<your-host>:8080/admin
```

建议立刻完成以下检查：

1. 能正常打开 `/admin`
2. 能使用日志中的初始密码登录
3. 创建一个网络并执行一次 `/knock`
4. 确认 `/app/data/doorman.db` 已生成且历史记录可见
5. 确认后台生成的客户端命令使用了正确的 `server.public_url`

## Release 二进制部署

不使用容器时，推荐使用 GitHub Release 二进制归档。Release 产物覆盖 Linux `amd64` 和 `arm64`，并且已经内嵌 Web 管理界面。

Release 资产文件名是：

- `doorman_${VERSION}_linux_amd64.tar.gz`
- `doorman_${VERSION}_linux_arm64.tar.gz`
- `SHA256SUMS`

### 下载并校验

先设置版本号和 CPU 架构，再下载 release 归档和校验文件：

```bash
VERSION=v0.2-alpha
ARCH=amd64

curl -LO "https://github.com/fishy-stick/doorman/releases/download/${VERSION}/doorman_${VERSION}_linux_${ARCH}.tar.gz"
curl -LO "https://github.com/fishy-stick/doorman/releases/download/${VERSION}/SHA256SUMS"
sha256sum --ignore-missing -c SHA256SUMS
```

ARM64 Linux 主机使用 `ARCH=arm64`。

解压归档：

```bash
tar -xzf "doorman_${VERSION}_linux_${ARCH}.tar.gz"
cd "doorman_${VERSION}_linux_${ARCH}"
```

解压后的目录包含：

- `doorman`
- `config.example.yaml`

### 准备文件和配置

典型生产目录结构：

- `/opt/doorman/doorman`
- `/opt/doorman/config.yaml`
- `/var/lib/doorman/doorman.db`

安装二进制和示例配置：

```bash
sudo install -d /opt/doorman /var/lib/doorman
sudo install -m 0755 doorman /opt/doorman/doorman
sudo install -m 0644 config.example.yaml /opt/doorman/config.yaml
```

编辑 `/opt/doorman/config.yaml`：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/var/lib/doorman/doorman.db"
  public_url: "https://your-domain.example"
```

`server.public_url` 是用于生成 `curl` 和 `crontab` 命令的外部访问地址。它可以包含路径前缀，例如 `https://www.abc.com/prefix`；这种情况下需要让反向代理把 `/prefix/knock` 转发到 Doorman 的 `/knock` 接口。

程序固定读取当前工作目录下的 `config.yaml`。如果 `server.db` 使用相对路径，例如 `doorman.db`，它也会相对于当前工作目录解析。生产环境更建议写绝对路径。

### 直接启动

直接在 shell 中运行时，需要从包含 `config.yaml` 的目录启动：

```bash
cd /opt/doorman
./doorman
```

确保运行 `./doorman` 的用户可以写入配置中的数据库目录，例如 `/var/lib/doorman`。

首次启动后查看标准输出日志，获取管理员初始密码。

## 使用 systemd 托管

长期运行的二进制部署建议使用专用服务用户：

```bash
sudo useradd --system --home /opt/doorman --shell /usr/sbin/nologin doorman
sudo chown -R doorman:doorman /var/lib/doorman
```

确认文件位于以下目录：

- `/opt/doorman/doorman`
- `/opt/doorman/config.yaml`
- `/var/lib/doorman/doorman.db`

使用下面的 `systemd` unit：

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

常用命令：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now doorman
sudo systemctl status doorman
sudo journalctl -u doorman -f
```

首次启动后，从 `journalctl` 中读取管理员初始密码。

## 从源码构建

源码构建适用于测试本地改动、调试构建，或目标环境没有合适 release 产物的情况。常规生产部署优先使用 GHCR 镜像或 GitHub Release 二进制。

构建机需要 Go `1.26+`、Node.js `22+` 和 `pnpm`。

先构建前端，再编译嵌入式二进制：

```bash
cd web
pnpm install
pnpm run build:embed
cd ..

go build -tags embedweb -o doorman ./cmd/doorman/
```

也可以在仓库根目录构建本地镜像：

```bash
docker build -t doorman:local .
```

源码构建产物仍使用上文相同的 `config.yaml`、数据持久化和 `server.public_url` 规则。

## 反向代理与 trust_proxy

如果 Doorman 部署在 Nginx、Caddy 或其他由你自己控制的反向代理后面，可以保留：

```yaml
server:
  trust_proxy: true
```

这时 Doorman 会优先信任：

1. `X-Forwarded-For`
2. `X-Real-IP`
3. `RemoteAddr`

如果服务直接对公网开放，或者前面的代理并不完全可信，建议关闭：

```yaml
server:
  trust_proxy: false
```

否则攻击者可能通过伪造请求头影响公网 IP 识别结果。

如果 `server.public_url` 带路径前缀，例如 `https://www.abc.com/prefix`，需要让反向代理把 `/prefix/knock` 转发到 Doorman 的 `/knock` 接口。

## 升级与变更注意事项

- GHCR 部署升级时，拉取新镜像标签并重建容器，同时保留原有数据卷或 bind mount
- Release 二进制升级时，替换 `/opt/doorman/doorman`，保留原有 `config.yaml` 和 SQLite 数据库
- 生产部署需要可重复回滚时建议固定具体版本标签；只有明确希望跟随最新稳定版时才使用 `latest`
- 修改 `server.public_url` 后，后台新生成的客户端命令会变化，已经复制到客户端的旧命令需要手动更新
- 重新生成网络 token 后，旧客户端命令会立即失效
- 管理员会话保存在内存中，服务重启后需要重新登录
- 修改 `trust_proxy` 前，先确认真实网络拓扑和代理链
