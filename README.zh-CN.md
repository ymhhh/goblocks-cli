# goblocks-cli

[English](README.md)

基于 [Goblocks](https://github.com/ymhhh/goblocks) 的脚手架 CLI，生成符合洋葱架构 / DDD 的 Go 服务工程。生成项目默认引用框架 **v0.3.1**。

## 安装

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
```

## 用法

```bash
goblocks new my-service --module github.com/acme/my-service
goblocks new demo-svc --module github.com/acme/demo-svc --demo
goblocks new full-svc --module github.com/acme/full-svc --demo --with-grpc --with-ai
```

| 参数 | 说明 |
|------|------|
| `--module` | Go module 路径（必填） |
| `--goblocks-version` | 框架版本（默认 `v0.3.1`） |
| `--demo` | User 示例（含 L2 挂载示例；L3 由 config `routes` 驱动） |
| `--with-grpc` | 生成 `proto/` 示例 |
| `--with-ai` | 生成 AI Chat handler |

生成后：

```bash
cd <output-dir>
go mod tidy
go run .
```

框架配置与分层限流说明见 [goblocks 文档](https://github.com/ymhhh/goblocks/tree/main/docs)（含 [中文文档](https://github.com/ymhhh/goblocks/tree/main/docs/zh)）。

## 限流与熔断（已有工程）

在工程根目录编辑 `config/config.yaml`（默认 `-f config/config.yaml`）：

```bash
goblocks config show

# L1 服务级全局限流（框架 app.Run 自动生效）
goblocks config global --rps 200 --burst 400

# L2 每用户默认配额（全接口共享；需在 infrastructure 挂载 UserRateLimit）
goblocks config user --enable --default-rps 30 --burst 60

# L3 单接口上限（配置 routes 后 app.Run 自动挂载；path 须与 Gin 路由模板一致）
goblocks config route add --method GET --path /users/:id --rps 50 --burst 100
goblocks config route add --method POST --path /ai/chat --rps 5 --burst 10
goblocks config route list
goblocks config route remove --method POST --path /ai/chat

# 熔断
goblocks config breaker --max-requests 5 --consecutive-failures 5

# 限流后端 memory | redis
goblocks config backend --backend redis --redis-addr redis://localhost:6379/0
```

## 开发

```bash
make build
make test
GOBLOCKS_PATH=/path/to/goblocks make test-integration
```

## License

GPL-3.0
