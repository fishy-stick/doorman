# Deploy

当前仓库提供的是基于 Docker 的部署方式，构建时会把前端静态资源嵌入到后端二进制里，运行时只需要一个容器。

## 文件说明

- 根目录 `Dockerfile`：多阶段构建镜像。
- 根目录 `.dockerignore`：避免把本地 `node_modules`、数据库和本地配置带进构建上下文。

## 构建镜像

在仓库根目录执行：

```bash
docker build -t doorman .
```

## 运行容器

最简单的启动方式：

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v doorman-data:/app/data \
  doorman
```

默认镜像内会生成一个 `/app/config.yaml`，内容等价于：

```yaml
server:
  port: 8080
  trust_proxy: true
  db: "/app/data/doorman.db"
```

这意味着：

- 服务监听 `8080`
- SQLite 数据库存放在 `/app/data/doorman.db`
- 只要挂载 `/app/data`，容器重建后数据仍会保留

## 使用自定义配置

如果你需要修改端口、数据库路径或 `trust_proxy`，可以准备自己的 `config.yaml`，挂载到容器内：

```bash
docker run -d \
  --name doorman \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  -v doorman-data:/app/data \
  doorman
```

如果你把数据库路径改到别的位置，记得同步挂载对应目录。

## 首次启动

首次启动时，服务会自动生成管理员密码并写到容器日志里。查看方式：

```bash
docker logs doorman
```

拿到密码后访问：

```text
http://<your-host>:8080/admin
```

建议首次登录后立即修改管理员密码。
