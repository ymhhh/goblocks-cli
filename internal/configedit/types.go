package configedit

// Root is the service config.yaml top-level shape (resilience edits preserve other keys).
type Root struct {
	Server     any              `yaml:"server,omitempty"`
	Resilience ResilienceConfig `yaml:"resilience"`
	AI         any              `yaml:"ai,omitempty"`
	Logger     any              `yaml:"logger,omitempty"`
	Metrics    any              `yaml:"metrics,omitempty"`
}

// ResilienceConfig matches goblocks resilience YAML schema.
type ResilienceConfig struct {
	Breaker   BreakerConfig   `yaml:"breaker"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// BreakerConfig holds circuit breaker settings.
type BreakerConfig struct {
	MaxRequests         uint32 `yaml:"max_requests"`
	ConsecutiveFailures uint32 `yaml:"consecutive_failures"`
	Interval            string `yaml:"interval"`
	Timeout             string `yaml:"timeout"`
}

// RateLimitConfig holds layered rate limiter settings.
type RateLimitConfig struct {
	Backend string                 `yaml:"backend"`
	Global  RateLimitTierConfig    `yaml:"global"`
	Redis   RedisRateLimitConfig   `yaml:"redis,omitempty"`
	User    UserRateLimitConfig    `yaml:"user"`
	Routes  []RouteRateLimitConfig `yaml:"routes,omitempty"`
}

// RateLimitTierConfig holds RPS/burst for one tier.
type RateLimitTierConfig struct {
	RPS   float64 `yaml:"rps"`
	Burst int     `yaml:"burst"`
}

// RedisRateLimitConfig holds Redis backend settings.
type RedisRateLimitConfig struct {
	Addr      string `yaml:"addr,omitempty"`
	KeyPrefix string `yaml:"key_prefix,omitempty"`
}

// UserRateLimitConfig holds per-user (L2) defaults.
type UserRateLimitConfig struct {
	Enabled    bool    `yaml:"enabled"`
	DefaultRPS float64 `yaml:"default_rps"`
	Burst      int     `yaml:"burst"`
}

// RouteRateLimitConfig holds per-route (L3) limits.
type RouteRateLimitConfig struct {
	Method string  `yaml:"method"`
	Path   string  `yaml:"path"`
	RPS    float64 `yaml:"rps"`
	Burst  int     `yaml:"burst"`
}
