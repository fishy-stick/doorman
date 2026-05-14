# 部署

语言： [English](README.md) | 简体中文

本文档覆盖 Doorman 的生产部署方式：

- Docker 部署，推荐优先使用
- 直接运行二进制
- `systemd` 托管

Doorman 推荐部署在家庭网络之外、具备公网可达性的服务器上，例如 VPS、云主机或其他长期在线的公网节点。生产形态是一个嵌入前端资源的 Go 二进制，运行时依赖 `config.yaml` 和 SQLite 数据文件。

## 部署前准备

部署前先确认：

- 服务器可以被目标家庭网络访问
- 已确定外部访问地址，并准备写入 `server.public_url`
- 已确定是否位于可信反向代理后面
- SQLite 数据文件所在目录会被持久化
- 防火墙或安全组已放行对外服务端口，默认是 `8080`

`server.public_url` 会影响后台生成的 `curl` 和 `crontab` 命令。正式部署时不要保留默认的 `http://your-server:8080`。

## Docker 部署

### 构建镜像

在仓库根目录执行：

```bash
docker build -t doorman .
```

镜像构建会：

- 使用 `web/` 构建前端静态资源
- 将前端产物嵌入 Go 二进制
- 在运行镜像中写入默认 `/app/config.yaml`

### 快速试跑

如果只是先确认服务可以启动，可以使用镜像内置配置：

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v doorman-data:/app/data \
  doorman
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
  doorman
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

## 直接运行二进制

### 构建

构建机需要 Go `1.26+`、Node.js `22+` 和 `pnpm`。

先构建前端，再编译嵌入式二进制：

```bash
cd web
pnpm install
pnpm run build:embed
cd ..

go build -tags embedweb -o doorman ./cmd/doorman/
```

### 准备配置

在二进制工作目录准备 `config.yaml`：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/var/lib/doorman/doorman.db"
  public_url: "https://your-domain.example"
```

`server.public_url` 是用于生成 `curl` 和 `crontab` 命令的外部访问地址。它可以包含路径前缀，例如 `https://www.abc.com/prefix`；这种情况下需要让反向代理把 `/prefix/knock` 转发到 Doorman 的 `/knock` 接口。

程序固定读取当前工作目录下的 `config.yaml`。如果你直接执行：

```bash
./doorman
```

那么当前 shell 所在目录必须包含 `config.yaml`。

如果 `server.db` 使用相对路径，例如 `doorman.db`，它也会相对于当前工作目录解析。生产环境更建议写绝对路径。

### 启动

```bash
./doorman
```

首次启动后查看标准输出日志，获取管理员初始密码。

## 使用 systemd 托管

约定目录结构：

- `/opt/doorman/doorman`
- `/opt/doorman/config.yaml`
- `/var/lib/doorman/doorman.db`

推荐做法：

- 为服务创建专用用户，例如 `doorman`
- 使用绝对数据库路径
- 确保 `WorkingDirectory` 指向包含 `config.yaml` 的目录
- 确保数据库目录对服务用户可写

下面是一个可直接调整的 `systemd` unit 示例：

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

- Docker 升级时，保留数据卷或 bind mount 中的数据库文件
- 二进制升级时，替换程序文件即可，保留原有 `config.yaml` 和 SQLite 数据库
- 修改 `server.public_url` 后，后台新生成的客户端命令会变化，已经复制到客户端的旧命令需要手动更新
- 重新生成网络 token 后，旧客户端命令会立即失效
- 管理员会话保存在内存中，服务重启后需要重新登录
- 修改 `trust_proxy` 前，先确认真实网络拓扑和代理链
