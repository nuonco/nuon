package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/actions/outputs"
)

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("supervisor uses POSIX shell scripts")
	}

	workdir := t.TempDir()
	outputFile := filepath.Join(workdir, outputs.Filename(0))

	t.Setenv(RootEnvVar, workdir)
	t.Setenv(OutputFilepathEnvVar, outputFile)

	t.Run("runs script and collects nuon_output", func(t *testing.T) {
		script := filepath.Join(workdir, "step.sh")
		contents := "#!/bin/sh\nnuon_output hello world\necho done\n"
		if err := os.WriteFile(script, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}

		code, err := Run(context.Background(), script, workdir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 0 {
			t.Fatalf("expected exit 0, got %d", code)
		}

		out, err := outputs.ParseFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}
		if out["hello"] != "world" {
			t.Fatalf("expected hello=world, got %v", out["hello"])
		}
	})

	t.Run("adds shebang when missing", func(t *testing.T) {
		script := filepath.Join(workdir, "no-shebang.sh")
		if err := os.WriteFile(script, []byte("echo hi\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		code, err := Run(context.Background(), script, workdir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 0 {
			t.Fatalf("expected exit 0, got %d", code)
		}

		contents, _ := os.ReadFile(script)
		if !strings.HasPrefix(string(contents), "#!") {
			t.Fatalf("expected shebang to be added, got %q", string(contents))
		}
	})

	t.Run("propagates non-zero exit", func(t *testing.T) {
		script := filepath.Join(workdir, "fail.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 7\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		code, err := Run(context.Background(), script, workdir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 7 {
			t.Fatalf("expected exit 7, got %d", code)
		}
	})
}
