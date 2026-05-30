# goblocks-cli

[Goblocks](https://github.com/ymhhh/goblocks) 脚手架 CLI，生成符合洋葱架构 / DDD 的 Go 服务工程。

## 安装

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
```

本地开发：

```bash
make build
./bin/goblocks --help
```

## 用法

```bash
goblocks new my-service --module github.com/acme/my-service
goblocks new demo-svc --module github.com/acme/demo-svc --demo
goblocks new full-svc --module github.com/acme/full-svc --demo --with-grpc --with-ai
```

### Flags

| Flag | 说明 |
|------|------|
| `--module` | Go module 路径（必填） |
| `--goblocks-version` | 生成工程引用的框架版本（默认 `v0.2.1`） |
| `--demo` | 生成 User Demo |
| `--with-grpc` | 额外生成 `proto/` 示例 |
| `--with-ai` | 生成 AI Chat handler |

生成工程的 `go.mod` 会声明：

```
require github.com/ymhhh/goblocks v0.2.1
```

本地联调框架时使用 `replace`：

```bash
echo "replace github.com/ymhhh/goblocks => /path/to/goblocks" >> go.mod
go mod tidy
```

## 版本对照

| CLI 版本 | 默认框架版本 | 说明 |
|----------|--------------|------|
| latest   | v0.2.1       | logger/config breaking change、可靠性改进 |

## 与 goblocks 框架的关系

| 仓库 | 职责 |
|------|------|
| [goblocks](https://github.com/ymhhh/goblocks) | 运行时框架库 |
| **goblocks-cli**（本仓库） | 项目脚手架与模板 |

## 开发

```bash
make test              # 单元测试
make test-integration  # 需设置 GOBLOCKS_PATH 指向本地 goblocks 仓库
```

```bash
export GOBLOCKS_PATH=/path/to/goblocks
make test-integration
```

模板位于 `internal/scaffold/templates/`（`empty` / `demo`）。

### 本地联调（go.work）

```bash
cd /path/to/github.com/ymhhh
go work init ./goblocks ./goblocks-cli
```

## License

GPL-3.0
