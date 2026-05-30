package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateAndBuildDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build integration in short mode")
	}

	goblocksRoot := os.Getenv("GOBLOCKS_PATH")
	if goblocksRoot == "" {
		t.Skip("GOBLOCKS_PATH not set, skipping build integration")
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")

	opts := Options{
		OutputDir:       out,
		ModulePath:      "github.com/acme/demo-svc",
		ServiceName:     "demo-svc",
		GoblocksVersion: DefaultGoblocksVersion,
		Demo:            true,
		WithGRPC:        true,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	goModPath := filepath.Join(out, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data) + "\nreplace github.com/ymhhh/goblocks => " + goblocksRoot + "\n"
	if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = out
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, outBytes)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "demo-svc-bin"), ".")
	cmd.Dir = out
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, outBytes)
	}
}
