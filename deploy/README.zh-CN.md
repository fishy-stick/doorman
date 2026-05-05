# 部署

语言： [English](README.md) | 简体中文

本文档覆盖 Doorman 的三种部署方式：

- Docker
- 直接运行二进制
- `systemd` 托管

Doorman 的生产形态是一个嵌入前端资源的 Go 二进制，运行时依赖 `config.yaml` 和 SQLite 数据文件。推荐将它部署在家庭之外、具备公网可达性的服务器上，例如 VPS、云主机或其他长期在线的公网节点。

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

### 默认运行方式

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
```

默认约束：

- 服务监听容器内 `8080`
- 数据库存放在 `/app/data/doorman.db`
- 运行用户是 `nobody:nogroup`
- 镜像声明了 `VOLUME /app/data`

只要 `/app/data` 做了持久化挂载，容器重建后数据库仍会保留。

### 使用自定义配置

如果你需要修改监听端口、数据库路径或 `trust_proxy`，可以挂载自己的配置文件：

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

## 直接运行二进制

### 构建

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
```

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

约定目录结构：

- `/opt/doorman/doorman`
- `/opt/doorman/config.yaml`
- `/var/lib/doorman/doorman.db`

推荐做法：

- 为服务创建专用用户，例如 `doorman`
- 使用绝对数据库路径
- 确保 `WorkingDirectory` 指向包含 `config.yaml` 的目录
- 确保数据库目录对服务用户可写

常用命令：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now doorman
sudo systemctl status doorman
sudo journalctl -u doorman -f
```

首次启动后，从 `journalctl` 中读取管理员初始密码。

## 反向代理与 trust_proxy

如果 Doorman 部署在 Nginx、Caddy 或其他反向代理后面，并且这些代理由你自己控制，可以保留：

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

## 升级与变更注意事项

- Docker 升级时，保留数据卷或 bind mount 中的数据库文件
- 重新生成网络 token 后，旧客户端命令会立即失效
- 管理员会话保存在内存中，服务重启后需要重新登录
- 修改 `trust_proxy` 前，先确认真实网络拓扑和代理链
