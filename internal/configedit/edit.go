package configedit

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/config.yaml"

// Load reads config.yaml.
func Load(path string) (*Root, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var root Root
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	normalize(&root)
	return &root, nil
}

// Save writes config.yaml.
func Save(path string, root *Root) error {
	normalize(root)
	data, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func normalize(root *Root) {
	if root.Resilience.Breaker.Interval == "" {
		root.Resilience.Breaker.Interval = "60s"
	}
	if root.Resilience.Breaker.Timeout == "" {
		root.Resilience.Breaker.Timeout = "30s"
	}
	if root.Resilience.Breaker.MaxRequests == 0 {
		root.Resilience.Breaker.MaxRequests = 3
	}
	if root.Resilience.Breaker.ConsecutiveFailures == 0 {
		root.Resilience.Breaker.ConsecutiveFailures = 3
	}
	if root.Resilience.RateLimit.Backend == "" {
		root.Resilience.RateLimit.Backend = "memory"
	}
	if root.Resilience.RateLimit.Global.RPS <= 0 {
		root.Resilience.RateLimit.Global.RPS = 100
	}
	if root.Resilience.RateLimit.Global.Burst <= 0 {
		root.Resilience.RateLimit.Global.Burst = 200
	}
	if root.Resilience.RateLimit.User.DefaultRPS <= 0 {
		root.Resilience.RateLimit.User.DefaultRPS = 20
	}
	if root.Resilience.RateLimit.User.Burst <= 0 {
		root.Resilience.RateLimit.User.Burst = 40
	}
}

// GlobalOpts updates L1 service-wide rate limits (framework app.Run).
type GlobalOpts struct {
	RPS   *float64
	Burst *int
}

func ApplyGlobal(root *Root, opts GlobalOpts) error {
	if opts.RPS != nil {
		if *opts.RPS <= 0 {
			return fmt.Errorf("global rps must be positive")
		}
		root.Resilience.RateLimit.Global.RPS = *opts.RPS
	}
	if opts.Burst != nil {
		if *opts.Burst <= 0 {
			return fmt.Errorf("global burst must be positive")
		}
		root.Resilience.RateLimit.Global.Burst = *opts.Burst
	}
	return nil
}

// UserOpts updates L2 per-user rate limits (mount UserRateLimit in infrastructure).
type UserOpts struct {
	Enabled    *bool
	DefaultRPS *float64
	Burst      *int
}

func ApplyUser(root *Root, opts UserOpts) error {
	if opts.Enabled != nil {
		root.Resilience.RateLimit.User.Enabled = *opts.Enabled
	}
	if opts.DefaultRPS != nil {
		if *opts.DefaultRPS <= 0 {
			return fmt.Errorf("user default_rps must be positive")
		}
		root.Resilience.RateLimit.User.DefaultRPS = *opts.DefaultRPS
	}
	if opts.Burst != nil {
		if *opts.Burst <= 0 {
			return fmt.Errorf("user burst must be positive")
		}
		root.Resilience.RateLimit.User.Burst = *opts.Burst
	}
	return nil
}

// RouteRule is one L3 API rate limit rule.
type RouteRule struct {
	Method string
	Path   string
	RPS    float64
	Burst  int
}

func AddRoute(root *Root, rule RouteRule) error {
	method := strings.ToUpper(strings.TrimSpace(rule.Method))
	path := strings.TrimSpace(rule.Path)
	if method == "" || path == "" {
		return fmt.Errorf("route method and path are required")
	}
	if rule.RPS <= 0 || rule.Burst <= 0 {
		return fmt.Errorf("route rps and burst must be positive")
	}
	entry := RouteRateLimitConfig{
		Method: method,
		Path:   path,
		RPS:    rule.RPS,
		Burst:  rule.Burst,
	}
	for i, r := range root.Resilience.RateLimit.Routes {
		if strings.EqualFold(r.Method, method) && r.Path == path {
			root.Resilience.RateLimit.Routes[i] = entry
			return nil
		}
	}
	root.Resilience.RateLimit.Routes = append(root.Resilience.RateLimit.Routes, entry)
	return nil
}

func RemoveRoute(root *Root, method, path string) (bool, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if method == "" || path == "" {
		return false, fmt.Errorf("route method and path are required")
	}
	routes := root.Resilience.RateLimit.Routes
	for i, r := range routes {
		if strings.EqualFold(r.Method, method) && r.Path == path {
			root.Resilience.RateLimit.Routes = append(routes[:i], routes[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// BreakerOpts updates circuit breaker settings.
type BreakerOpts struct {
	MaxRequests         *uint32
	ConsecutiveFailures *uint32
	Interval            *string
	Timeout             *string
}

func ApplyBreaker(root *Root, opts BreakerOpts) error {
	if opts.MaxRequests != nil {
		if *opts.MaxRequests == 0 {
			return fmt.Errorf("max_requests must be positive")
		}
		root.Resilience.Breaker.MaxRequests = *opts.MaxRequests
	}
	if opts.ConsecutiveFailures != nil {
		if *opts.ConsecutiveFailures == 0 {
			return fmt.Errorf("consecutive_failures must be positive")
		}
		root.Resilience.Breaker.ConsecutiveFailures = *opts.ConsecutiveFailures
	}
	if opts.Interval != nil {
		root.Resilience.Breaker.Interval = *opts.Interval
	}
	if opts.Timeout != nil {
		root.Resilience.Breaker.Timeout = *opts.Timeout
	}
	return nil
}

// BackendOpts updates rate limit backend (memory | redis).
type BackendOpts struct {
	Backend   *string
	RedisAddr *string
	KeyPrefix *string
}

func ApplyBackend(root *Root, opts BackendOpts) error {
	if opts.Backend != nil {
		b := strings.ToLower(strings.TrimSpace(*opts.Backend))
		if b != "memory" && b != "redis" {
			return fmt.Errorf("backend must be memory or redis")
		}
		root.Resilience.RateLimit.Backend = b
		if b == "redis" && (opts.RedisAddr == nil || strings.TrimSpace(*opts.RedisAddr) == "") {
			if strings.TrimSpace(root.Resilience.RateLimit.Redis.Addr) == "" {
				return fmt.Errorf("redis backend requires --redis-addr or existing resilience.rate_limit.redis.addr")
			}
		}
	}
	if opts.RedisAddr != nil {
		root.Resilience.RateLimit.Redis.Addr = strings.TrimSpace(*opts.RedisAddr)
	}
	if opts.KeyPrefix != nil {
		root.Resilience.RateLimit.Redis.KeyPrefix = strings.TrimSpace(*opts.KeyPrefix)
	}
	return nil
}

// FormatResilience returns a human-readable summary.
func FormatResilience(r ResilienceConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "breaker:\n")
	fmt.Fprintf(&b, "  max_requests: %d\n", r.Breaker.MaxRequests)
	fmt.Fprintf(&b, "  consecutive_failures: %d\n", r.Breaker.ConsecutiveFailures)
	fmt.Fprintf(&b, "  interval: %s\n", r.Breaker.Interval)
	fmt.Fprintf(&b, "  timeout: %s\n", r.Breaker.Timeout)
	fmt.Fprintf(&b, "rate_limit:\n")
	fmt.Fprintf(&b, "  backend: %s\n", r.RateLimit.Backend)
	if r.RateLimit.Redis.Addr != "" {
		fmt.Fprintf(&b, "  redis.addr: %s\n", r.RateLimit.Redis.Addr)
	}
	if r.RateLimit.Redis.KeyPrefix != "" {
		fmt.Fprintf(&b, "  redis.key_prefix: %s\n", r.RateLimit.Redis.KeyPrefix)
	}
	fmt.Fprintf(&b, "  global (L1 服务级): rps=%.0f burst=%d\n", r.RateLimit.Global.RPS, r.RateLimit.Global.Burst)
	fmt.Fprintf(&b, "  user (L2 每用户配额): enabled=%v default_rps=%.0f burst=%d\n",
		r.RateLimit.User.Enabled, r.RateLimit.User.DefaultRPS, r.RateLimit.User.Burst)
	if r.RateLimit.User.Enabled {
		fmt.Fprintf(&b, "    → 需在 infrastructure 挂载 UserRateLimit + 注入 userId\n")
	}
	if len(r.RateLimit.Routes) == 0 {
		fmt.Fprintf(&b, "  routes (L3 按 API): (none)\n")
	} else {
		fmt.Fprintf(&b, "  routes (L3 按 API):\n")
		for _, rt := range r.RateLimit.Routes {
			fmt.Fprintf(&b, "    - %s %s: rps=%.0f burst=%d\n", rt.Method, rt.Path, rt.RPS, rt.Burst)
		}
	}
	return b.String()
}
