# Doorman

[![Version](https://img.shields.io/github/v/tag/fishy-stick/doorman?sort=semver&filter=%21%2A-%2A&label=version)](https://github.com/fishy-stick/doorman/tags)
[![Docker Publish](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/fishy-stick/doorman/actions/workflows/docker-publish.yml)

语言： [English](README.md) | 简体中文

Doorman 是一个面向家庭动态公网 IP 场景的轻量级 DDNS 服务端。它通常部署在家庭之外的公网服务器上，接收来自家庭内网客户端的 HTTP 请求，并在识别到公网 IP 变化时触发 DDNS 更新。

项目自带管理后台和 API，生产环境可以打包成单一二进制运行；数据保存在 SQLite 中，不依赖额外数据库。

## 功能特性

- 支持多个内网，每个内网拥有独立名称、token 和 DDNS 配置
- 根据请求来源自动识别公网 IP
- 支持 `X-Forwarded-For`、`X-Real-IP` 和 `RemoteAddr`
- 使用 Bearer Token 保护 `/knock` 接口
- 自动记录 IP 历史、变化状态和 DDNS 执行结果
- 当前内置 `DNSPod` Provider
- 管理后台可直接生成 `curl` 和 `crontab` 命令
- 使用 SQLite 持久化数据，重启后记录不丢失

## 环境要求

- Go `1.26+`
- Node.js `22+`
- `pnpm`

如果本机没有启用 `pnpm`，先执行：

```bash
corepack enable
```

## 快速开始

### 开发模式

开发模式下，后端和前端分别启动：

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

访问 `http://127.0.0.1:15173/admin/`。

说明：

- `go run ./cmd/doorman/` 需要在包含 `config.yaml` 的目录执行
- 开发模式下管理界面由 Vite 提供，后端只负责 `/admin/api` 和 `/knock`
- 首次启动会在日志中输出管理员密码

### 生产构建

生产环境需要先构建前端静态资源，再将其嵌入 Go 二进制：

```bash
cd web
pnpm install
pnpm run build:embed
cd ..

go build -tags embedweb -o doorman ./cmd/doorman/
./doorman
```

访问 `http://127.0.0.1:8080/admin`。

如果不带 `embedweb` 构建标签，二进制不会内置管理界面，`/admin` 也不会提供前端页面。

## 配置

运行配置通过 `config.yaml` 提供，内网和 DDNS 规则通过 Web 管理界面维护。

示例：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "doorman.db"
  public_url: "http://your-server:8080"
```

字段说明：

| 配置项 | 说明 |
|--------|------|
| `server.port` | 监听端口，默认 `:8080`。可写 `8080` 或 `:8080`，程序会自动归一化。 |
| `server.trust_proxy` | 是否信任代理头，默认 `true`。 |
| `server.db` | SQLite 数据库文件路径，默认 `doorman.db`。 |
| `server.public_url` | 用于生成客户端命令的外部访问地址，默认 `http://your-server:8080`。可以带路径前缀，例如 `https://www.abc.com/prefix`。不支持 query 和 fragment。 |

## 使用流程

1. 启动服务并查看日志中的初始管理员密码。
2. 访问 `/admin` 登录后台。
3. 创建一个网络。
4. 复制系统生成的 `curl` 或 `crontab` 命令。
5. 在目标内网中定时调用公网服务器上的 `/knock`。
6. 在管理后台查看当前 IP、历史记录和 DDNS 状态。

创建网络时，Doorman 会自动生成 Bearer Token。重新生成 token 后，旧客户端命令会立刻失效，需要同步更新。

## 客户端调用

客户端只需要发起标准 HTTP 请求：

```bash
curl -H "Authorization: Bearer your-token" http://your-server:8080/knock
```

管理后台会根据 `server.public_url` 生成这条命令。如果配置为 `https://www.abc.com/prefix`，生成的目标地址是 `https://www.abc.com/prefix/knock`；此时需要让反向代理把该路径转发到 Doorman 的 `/knock` 接口。

返回结果会包含当前识别到的 IP、是否发生变化，以及本次是否执行了 DDNS 更新。

## 运行行为

### IP 识别规则

当 `server.trust_proxy=true` 时，Doorman 按下面顺序取客户端 IP：

1. `X-Forwarded-For` 中第一个合法 IP
2. `X-Real-IP`
3. `RemoteAddr`

如果服务直接暴露在公网、前面没有你自己控制的可信反向代理，建议将 `trust_proxy` 设为 `false`，避免伪造请求头影响 IP 识别。

### 会话与登录

- 管理员会话通过 Cookie 保存在内存中
- 单次会话有效期为 24 小时
- 服务重启后，所有后台会话都会失效
- 修改管理员密码后，当前所有登录会话会被强制清空

### DDNS 执行规则

- 只有网络启用了 DDNS，且本次识别到的 IP 与上次不同，才会触发更新
- 当前仅支持 `DNSPod`
- 即使未启用 DDNS，Doorman 仍会记录 `/knock` 历史

## 部署

支持 Docker、直接运行二进制和 `systemd` 托管，详见 [deploy/README.zh-CN.md](deploy/README.zh-CN.md)。

## 设计原则

- **不信任客户端上报的 IP**：只根据请求来源推断公网地址
- **尽量简单**：单二进制、SQLite、少依赖
- **面向自用场景**：优先覆盖家庭网络中的动态公网 IP 管理需求

## License

MIT
