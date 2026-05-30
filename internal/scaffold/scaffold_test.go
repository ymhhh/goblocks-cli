package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateEmpty(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "testsvc")

	opts := Options{
		OutputDir:  out,
		ModulePath: "github.com/acme/testsvc",
		ServiceName: "testsvc",
		Demo:       false,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "main.go")
	checkFile(t, out, "go.mod")
	checkFile(t, out, "config/config.yaml")
	checkConfigContains(t, out, "logger:")
	checkConfigContains(t, out, "rate_limit:")
	checkConfigContains(t, out, "global:")
	checkConfigContains(t, out, "user:")
	checkConfigContains(t, out, "health:")
	checkConfigNotContains(t, out, "rate_limit:\n    rps:")
	checkGoModGoblocksVersion(t, out, DefaultGoblocksVersion)
	checkFile(t, out, "infrastructure/run.go")
	checkFileNotContains(t, out, "infrastructure/run.go", `engine.GET("/health"`)
	checkFile(t, out, "infrastructure/grpc_server.go")
	checkNoFile(t, out, "core/user.go")
}

func TestGenerateDemoWithGRPCProto(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")

	opts := Options{
		OutputDir:   out,
		ModulePath:  "github.com/acme/demo-svc",
		ServiceName: "demo-svc",
		Demo:        true,
		WithGRPC:    true,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "proto/user/v1/user.proto")
}

func TestGenerateDemo(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")

	opts := Options{
		OutputDir:  out,
		ModulePath: "github.com/acme/demo-svc",
		ServiceName: "demo-svc",
		Demo:       true,
		WithGRPC:   true,
		WithAI:     true,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "core/user.go")
	checkFile(t, out, "handlers/user_handler.go")
	checkFile(t, out, "handlers/ai_handler.go")
	checkFile(t, out, "proto/user/v1/user.proto")
	checkFile(t, out, "infrastructure/grpc_server.go")
	checkConfigContains(t, out, "user:")
	checkConfigContains(t, out, "routes:")
	checkConfigContains(t, out, "/users/:id")
	checkFileContains(t, out, "infrastructure/run.go", "UserRateLimit")
	checkFileContains(t, out, "infrastructure/run.go", "rate_limit.routes")
	checkFileNotContains(t, out, "infrastructure/run.go", `engine.GET("/health"`)
}

func checkFile(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", rel, err)
	}
}

func checkNoFile(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no file %s", rel)
	}
}

func checkConfigContains(t *testing.T, root, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "config/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("config/config.yaml missing %q", want)
	}
}

func checkConfigNotContains(t *testing.T, root, unwanted string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "config/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), unwanted) {
		t.Fatalf("config/config.yaml must not contain %q", unwanted)
	}
}

func checkGoModGoblocksVersion(t *testing.T, root, version string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	want := "github.com/ymhhh/goblocks " + version
	if !strings.Contains(string(data), want) {
		t.Fatalf("go.mod missing %q", want)
	}
}

func checkFileContains(t *testing.T, root, rel, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s missing %q", rel, want)
	}
}

func checkFileNotContains(t *testing.T, root, rel, unwanted string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), unwanted) {
		t.Fatalf("%s must not contain %q", rel, unwanted)
	}
}
