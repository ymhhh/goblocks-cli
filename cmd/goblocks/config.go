package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ymhhh/goblocks-cli/internal/configedit"
)

func configCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Edit resilience rate limit and circuit breaker settings",
		Long: `Patch config/config.yaml for goblocks resilience settings.

Layers:
  L1 global  — service-wide (mounted by app.Run)
  L2 user    — per-user default (mount httpmiddleware.UserRateLimit in infrastructure)
  L3 routes  — per API method+path (app.Run auto-mounts when routes are in config)`,
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "f", configedit.DefaultConfigPath, "Path to config.yaml")

	cmd.AddCommand(
		configShowCmd(&configPath),
		configGlobalCmd(&configPath),
		configUserCmd(&configPath),
		configRouteCmd(&configPath),
		configBreakerCmd(&configPath),
		configBackendCmd(&configPath),
	)

	return cmd
}

func configShowCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current resilience settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := configedit.Load(*configPath)
			if err != nil {
				return err
			}
			fmt.Print(configedit.FormatResilience(root.Resilience))
			return nil
		},
	}
}

func configGlobalCmd(configPath *string) *cobra.Command {
	var rps float64
	var burst int

	cmd := &cobra.Command{
		Use:   "global",
		Short: "Set L1 service-wide rate limit (global rps/burst)",
		Long:  "Updates resilience.rate_limit.global. Applied automatically by framework app.Run.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("rps") && !cmd.Flags().Changed("burst") {
				return fmt.Errorf("specify at least one of --rps or --burst")
			}
			return patchConfig(*configPath, func(root *configedit.Root) error {
				opts := configedit.GlobalOpts{}
				if cmd.Flags().Changed("rps") {
					opts.RPS = &rps
				}
				if cmd.Flags().Changed("burst") {
					opts.Burst = &burst
				}
				return configedit.ApplyGlobal(root, opts)
			}, "updated L1 global rate limit")
		},
	}
	cmd.Flags().Float64Var(&rps, "rps", 0, "Global requests per second")
	cmd.Flags().IntVar(&burst, "burst", 0, "Global burst capacity")
	return cmd
}

func configUserCmd(configPath *string) *cobra.Command {
	var (
		enable  bool
		disable bool
		rps     float64
		burst   int
	)

	cmd := &cobra.Command{
		Use:   "user",
		Short: "Set L2 per-user rate limit (user.default_rps/burst)",
		Long: `Updates resilience.rate_limit.user (L2 每用户默认配额).

与 L3 单接口限流配合：L2 限制同一用户在所有接口上的总配额；对用户 API 还可通过
"goblocks config route add" 配置 routes（如 GET /users/:id），由 app.Run 自动挂载 L3。

启用 L2 后需在 infrastructure/registerHTTP 挂载 UserRateLimit，并在鉴权后注入 userId
（demo 模板在 GET /users/:id 链上示例）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rpsChanged := cmd.Flags().Changed("default-rps")
			if !cmd.Flags().Changed("enable") &&
				!cmd.Flags().Changed("disable") &&
				!rpsChanged &&
				!cmd.Flags().Changed("burst") {
				return fmt.Errorf("specify at least one of --enable, --disable, --default-rps, or --burst")
			}
			if cmd.Flags().Changed("enable") && cmd.Flags().Changed("disable") {
				return fmt.Errorf("cannot use both --enable and --disable")
			}
			return patchConfig(*configPath, func(root *configedit.Root) error {
				opts := configedit.UserOpts{}
				if cmd.Flags().Changed("enable") {
					v := true
					opts.Enabled = &v
				}
				if cmd.Flags().Changed("disable") {
					v := false
					opts.Enabled = &v
				}
				if rpsChanged {
					opts.DefaultRPS = &rps
				}
				if cmd.Flags().Changed("burst") {
					opts.Burst = &burst
				}
				if err := configedit.ApplyUser(root, opts); err != nil {
					return err
				}
				if root.Resilience.RateLimit.User.Enabled {
					fmt.Fprintln(os.Stderr, "hint: mount UserRateLimit after GinContextWithUserID (see demo infrastructure/run.go)")
				}
				return nil
			}, "updated L2 user rate limit")
		},
	}
	cmd.Flags().BoolVar(&enable, "enable", false, "Enable L2 user rate limiting")
	cmd.Flags().BoolVar(&disable, "disable", false, "Disable L2 user rate limiting")
	cmd.Flags().Float64Var(&rps, "default-rps", 0, "Per-user default RPS (maps to user.default_rps)")
	cmd.Flags().IntVar(&burst, "burst", 0, "Per-user burst (maps to user.burst)")
	return cmd
}

func configRouteCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Manage L3 per-API rate limit rules",
	}

	var (
		method string
		path   string
		rps    float64
		burst  int
	)

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add or update an API rate limit rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return patchConfig(*configPath, func(root *configedit.Root) error {
				return configedit.AddRoute(root, configedit.RouteRule{
					Method: method,
					Path:   path,
					RPS:    rps,
					Burst:  burst,
				})
			}, fmt.Sprintf("updated route %s %s", method, path))
		},
	}
	addCmd.Flags().StringVar(&method, "method", "", "HTTP method (GET, POST, ...)")
	addCmd.Flags().StringVar(&path, "path", "", "Route path (e.g. /ai/chat)")
	addCmd.Flags().Float64Var(&rps, "rps", 0, "Requests per second for this API")
	addCmd.Flags().IntVar(&burst, "burst", 0, "Burst for this API")
	_ = addCmd.MarkFlagRequired("method")
	_ = addCmd.MarkFlagRequired("path")
	_ = addCmd.MarkFlagRequired("rps")
	_ = addCmd.MarkFlagRequired("burst")

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an API rate limit rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return patchConfig(*configPath, func(root *configedit.Root) error {
				ok, err := configedit.RemoveRoute(root, method, path)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("route not found: %s %s", method, path)
				}
				return nil
			}, fmt.Sprintf("removed route %s %s", method, path))
		},
	}
	removeCmd.Flags().StringVar(&method, "method", "", "HTTP method")
	removeCmd.Flags().StringVar(&path, "path", "", "Route path")
	_ = removeCmd.MarkFlagRequired("method")
	_ = removeCmd.MarkFlagRequired("path")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List L3 API rate limit rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := configedit.Load(*configPath)
			if err != nil {
				return err
			}
			routes := root.Resilience.RateLimit.Routes
			if len(routes) == 0 {
				fmt.Println("(no routes)")
				return nil
			}
			for _, r := range routes {
				fmt.Printf("%s %s  rps=%.0f burst=%d\n", r.Method, r.Path, r.RPS, r.Burst)
			}
			return nil
		},
	}

	cmd.AddCommand(addCmd, removeCmd, listCmd)
	return cmd
}

func configBreakerCmd(configPath *string) *cobra.Command {
	var (
		maxRequests         uint32
		consecutiveFailures uint32
		interval            string
		timeout             string
	)

	cmd := &cobra.Command{
		Use:   "breaker",
		Short: "Set circuit breaker parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("max-requests") &&
				!cmd.Flags().Changed("consecutive-failures") &&
				!cmd.Flags().Changed("interval") &&
				!cmd.Flags().Changed("timeout") {
				return fmt.Errorf("specify at least one breaker flag")
			}
			return patchConfig(*configPath, func(root *configedit.Root) error {
				opts := configedit.BreakerOpts{}
				if cmd.Flags().Changed("max-requests") {
					opts.MaxRequests = &maxRequests
				}
				if cmd.Flags().Changed("consecutive-failures") {
					opts.ConsecutiveFailures = &consecutiveFailures
				}
				if cmd.Flags().Changed("interval") {
					opts.Interval = &interval
				}
				if cmd.Flags().Changed("timeout") {
					opts.Timeout = &timeout
				}
				return configedit.ApplyBreaker(root, opts)
			}, "updated circuit breaker")
		},
	}
	cmd.Flags().Uint32Var(&maxRequests, "max-requests", 3, "Max requests allowed in half-open state")
	cmd.Flags().Uint32Var(&consecutiveFailures, "consecutive-failures", 3, "Failures before opening breaker")
	cmd.Flags().StringVar(&interval, "interval", "60s", "Statistics interval (e.g. 60s)")
	cmd.Flags().StringVar(&timeout, "timeout", "30s", "Open state duration before half-open")
	return cmd
}

func configBackendCmd(configPath *string) *cobra.Command {
	var (
		backend   string
		redisAddr string
		keyPrefix string
	)

	cmd := &cobra.Command{
		Use:   "backend",
		Short: "Set rate limit backend (memory or redis)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("backend") &&
				!cmd.Flags().Changed("redis-addr") &&
				!cmd.Flags().Changed("key-prefix") {
				return fmt.Errorf("specify at least one of --backend, --redis-addr, --key-prefix")
			}
			return patchConfig(*configPath, func(root *configedit.Root) error {
				opts := configedit.BackendOpts{}
				if cmd.Flags().Changed("backend") {
					opts.Backend = &backend
				}
				if cmd.Flags().Changed("redis-addr") {
					opts.RedisAddr = &redisAddr
				}
				if cmd.Flags().Changed("key-prefix") {
					opts.KeyPrefix = &keyPrefix
				}
				return configedit.ApplyBackend(root, opts)
			}, "updated rate limit backend")
		},
	}
	cmd.Flags().StringVar(&backend, "backend", "memory", "Backend: memory or redis")
	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address (required when backend=redis)")
	cmd.Flags().StringVar(&keyPrefix, "key-prefix", "", "Redis key prefix (default goblocks:rl:)")
	return cmd
}

func patchConfig(path string, patch func(*configedit.Root) error, okMsg string) error {
	root, err := configedit.Load(path)
	if err != nil {
		return err
	}
	if err := patch(root); err != nil {
		return err
	}
	if err := configedit.Save(path, root); err != nil {
		return err
	}
	fmt.Printf("%s in %s\n", okMsg, path)
	return nil
}
