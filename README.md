# goblocks-cli

[中文文档](README.zh-CN.md)

CLI to scaffold Go services with the [Goblocks](https://github.com/ymhhh/goblocks) framework using onion architecture / DDD. Generated projects default to framework **v0.3.1**.

## Install

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
```

## Usage

```bash
goblocks new my-service --module github.com/acme/my-service
goblocks new demo-svc --module github.com/acme/demo-svc --demo
goblocks new full-svc --module github.com/acme/full-svc --demo --with-grpc --with-ai
```

| Flag | Description |
|------|-------------|
| `--module` | Go module path (required) |
| `--goblocks-version` | Framework version (default `v0.3.1`) |
| `--demo` | User demo (L2 wiring example; L3 driven by config `routes`) |
| `--with-grpc` | Include `proto/` sample |
| `--with-ai` | Include AI chat handler |

After generation:

```bash
cd <output-dir>
go mod tidy
go run .
```

See [Goblocks docs](https://github.com/ymhhh/goblocks/tree/main/docs) for framework configuration and layered rate limiting.

## Rate limits & circuit breaker (existing projects)

Edit `config/config.yaml` from the project root (default `-f config/config.yaml`):

```bash
goblocks config show

# L1 service-wide (applied by app.Run)
goblocks config global --rps 200 --burst 400

# L2 per-user default quota (mount UserRateLimit in infrastructure)
goblocks config user --enable --default-rps 30 --burst 60

# L3 per-route cap (auto-mounted when routes exist; path must match Gin route template)
goblocks config route add --method GET --path /users/:id --rps 50 --burst 100
goblocks config route add --method POST --path /ai/chat --rps 5 --burst 10
goblocks config route list
goblocks config route remove --method POST --path /ai/chat

# Circuit breaker
goblocks config breaker --max-requests 5 --consecutive-failures 5

# Rate limit backend: memory | redis
goblocks config backend --backend redis --redis-addr redis://localhost:6379/0
```

## Development

```bash
make build
make test
GOBLOCKS_PATH=/path/to/goblocks make test-integration
```

## License

GPL-3.0
