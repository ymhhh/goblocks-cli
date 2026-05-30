package configedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleYAML = `server:
  http:
    addr: ":8080"
resilience:
  breaker:
    max_requests: 3
    consecutive_failures: 3
    interval: 60s
    timeout: 30s
  rate_limit:
    backend: memory
    global:
      rps: 100
      burst: 200
    user:
      enabled: false
      default_rps: 20
      burst: 40
    routes:
      - method: POST
        path: /ai/chat
        rps: 5
        burst: 10
`

func writeSample(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(sampleYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyGlobal(t *testing.T) {
	path := writeSample(t)
	root, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	rps := 50.0
	burst := 80
	if err := ApplyGlobal(root, GlobalOpts{RPS: &rps, Burst: &burst}); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, root); err != nil {
		t.Fatal(err)
	}
	root2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if root2.Resilience.RateLimit.Global.RPS != 50 || root2.Resilience.RateLimit.Global.Burst != 80 {
		t.Fatalf("global: got rps=%v burst=%v", root2.Resilience.RateLimit.Global.RPS, root2.Resilience.RateLimit.Global.Burst)
	}
}

func TestApplyUser(t *testing.T) {
	path := writeSample(t)
	root, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	enabled := true
	rps := 10.0
	burst := 25
	if err := ApplyUser(root, UserOpts{Enabled: &enabled, DefaultRPS: &rps, Burst: &burst}); err != nil {
		t.Fatal(err)
	}
	u := root.Resilience.RateLimit.User
	if !u.Enabled || u.DefaultRPS != 10 || u.Burst != 25 {
		t.Fatalf("user: %+v", u)
	}
}

func TestAddAndRemoveRoute(t *testing.T) {
	path := writeSample(t)
	root, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := AddRoute(root, RouteRule{Method: "GET", Path: "/users/:id", RPS: 30, Burst: 60}); err != nil {
		t.Fatal(err)
	}
	if err := AddRoute(root, RouteRule{Method: "POST", Path: "/ai/chat", RPS: 8, Burst: 16}); err != nil {
		t.Fatal(err)
	}
	if len(root.Resilience.RateLimit.Routes) != 2 {
		t.Fatalf("routes len=%d", len(root.Resilience.RateLimit.Routes))
	}
	found := false
	for _, r := range root.Resilience.RateLimit.Routes {
		if r.Method == "POST" && r.Path == "/ai/chat" && r.RPS == 8 {
			found = true
		}
	}
	if !found {
		t.Fatal("upsert route failed")
	}
	ok, err := RemoveRoute(root, "GET", "/users/:id")
	if err != nil || !ok {
		t.Fatalf("remove: ok=%v err=%v", ok, err)
	}
	if len(root.Resilience.RateLimit.Routes) != 1 {
		t.Fatalf("after remove len=%d", len(root.Resilience.RateLimit.Routes))
	}
}

func TestApplyBreaker(t *testing.T) {
	path := writeSample(t)
	root, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	max := uint32(5)
	interval := "90s"
	if err := ApplyBreaker(root, BreakerOpts{MaxRequests: &max, Interval: &interval}); err != nil {
		t.Fatal(err)
	}
	if root.Resilience.Breaker.MaxRequests != 5 || root.Resilience.Breaker.Interval != "90s" {
		t.Fatalf("breaker: %+v", root.Resilience.Breaker)
	}
}

func TestFormatResilience(t *testing.T) {
	path := writeSample(t)
	root, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatResilience(root.Resilience)
	if !strings.Contains(out, "L1") || !strings.Contains(out, "POST /ai/chat") {
		t.Fatalf("format:\n%s", out)
	}
}
