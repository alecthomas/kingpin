package kingpin

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDeadCodeElimination verifies that programs using kingpin's default
// UsageRenderer do not link in reflect.MethodByName, which is the key
// indicator that dead code elimination is working.
func TestDeadCodeElimination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("go tool nm not reliable on Windows")
	}

	dir := t.TempDir()
	filename := filepath.Join(dir, "main.go")
	err := os.WriteFile(filename, []byte(`package main

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

func main() {
	app := kingpin.New("test", "A test app.")
	app.UsageRenderer(kingpin.RenderDefault)
	app.Flag("verbose", "Enable verbose mode.").Bool()
	app.Command("sub", "A subcommand.")
	app.Parse(os.Args[1:])
}
`), 0o600)
	require.NoError(t, err)

	binPath := filepath.Join(dir, "test_binary")
	buf, err := exec.Command("go", "build", "-trimpath", "-o", binPath, filename).CombinedOutput()
	require.NoError(t, err, "go build failed: %s", buf)

	buf, err = exec.Command("go", "tool", "nm", binPath).CombinedOutput()
	require.NoError(t, err, "go tool nm failed: %s", buf)

	require.NotContains(t, string(buf), "MethodByName", "text/template was not eliminated by dead code elimination")
}
