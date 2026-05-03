# Doorman

轻量级 DDNS 服务端工具，适用于家庭动态公网 IP 场景，运行于网关之后。

自带前后端，单一二进制部署。通过 HTTP 接口接收客户端请求，基于请求来源识别公网 IP，按配置条件触发 DDNS 更新。支持多个内网，每个内网独立配置。

## 功能特性

- 支持多内网，每个内网独立 token 和 DDNS 配置
- 从请求来源自动获取公网 IP（支持 X-Forwarded-For、X-Real-IP、RemoteAddr）
- 基于 Bearer Token 的请求认证
- IP 变化检测与历史记录
- DDNS Provider 可扩展（当前支持 DNSPod）
- 自动生成 curl / crontab 调用命令
- SQLite 持久化，重启不丢失数据
- 单一二进制部署，无外部依赖

## 快速开始

```bash
# 安装前端依赖
cd web && pnpm install && cd ..

cp config.example.yaml config.yaml
# 编辑 config.yaml 配置端口、数据库路径

# 开发：后端 API
go run ./cmd/doorman/

# 开发：前端 dev server（反代 /admin/api 和 /knock 到后端）
cd web && pnpm dev && cd ..
# 访问 http://127.0.0.1:5173/admin/
```

```bash
# 生产构建：前端产物输出到 internal/webui/dist
cd web && pnpm run build:embed && cd ..

# 生产构建：仅正式打包时嵌入前端
go build -tags embedweb -o doorman ./cmd/doorman/

./doorman
# 访问 http://localhost:8080/admin
```

## 配置

服务运行参数通过 YAML 配置，内网和 DDNS 通过 Web 管理：

| 配置项 | 说明 |
|--------|------|
| `server.port` | 监听端口，默认 `:8080` |
| `server.trust_proxy` | 是否信任代理头，默认 `true` |
| `server.db` | SQLite 数据库文件路径，默认 `doorman.db` |

首次启动时自动生成随机管理员密码，请查看日志输出，并在首次登录后及时修改。

## 客户端调用

通过标准 HTTP 请求调用，可结合 crontab 或 systemd timer 定时执行：

```bash
curl -H "Authorization: Bearer your-token" http://your-server:8080/knock
```

## 部署

支持直接运行、Docker、systemd 等方式部署，详见 [deploy/](deploy/) 目录。

## 设计原则

- **不信任客户端**：IP 始终从请求来源获取，不依赖请求体
- **极简设计**：单一二进制，SQLite 存储，无外部依赖
- **自用工具**：面向家庭网络场景，不追求高可用

## License

MIT
