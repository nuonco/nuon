package supervisor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/actions/outputs"
)

// runSupervisor writes the embedded script into workdir and runs it via /bin/sh,
// mirroring how the launcher invokes it inside the container.
func runSupervisor(t *testing.T, workdir, scriptPath, outputFile string) int {
	t.Helper()

	supPath, err := Write(workdir)
	if err != nil {
		t.Fatalf("write supervisor: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "/bin/sh", supPath, "--script", scriptPath, "--workdir", workdir)
	cmd.Env = append(os.Environ(),
		RootEnvVar+"="+workdir,
		OutputFilepathEnvVar+"="+outputFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := asExit(err, &exitErr); ok {
			return exitErr.ExitCode()
		}
		t.Fatalf("run supervisor: %v", err)
	}
	return 0
}

func asExit(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

func TestSupervisor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("supervisor is a POSIX shell script")
	}

	t.Run("runs script and collects nuon_output", func(t *testing.T) {
		workdir := t.TempDir()
		outputFile := filepath.Join(workdir, outputs.Filename(0))
		script := filepath.Join(workdir, "step.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\nnuon_output hello world\necho done\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if code := runSupervisor(t, workdir, script, outputFile); code != 0 {
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
		workdir := t.TempDir()
		outputFile := filepath.Join(workdir, outputs.Filename(0))
		script := filepath.Join(workdir, "no-shebang.sh")
		if err := os.WriteFile(script, []byte("echo hi\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if code := runSupervisor(t, workdir, script, outputFile); code != 0 {
			t.Fatalf("expected exit 0, got %d", code)
		}

		contents, _ := os.ReadFile(script)
		if !strings.HasPrefix(string(contents), "#!") {
			t.Fatalf("expected shebang to be added, got %q", string(contents))
		}
	})

	t.Run("propagates non-zero exit", func(t *testing.T) {
		workdir := t.TempDir()
		outputFile := filepath.Join(workdir, outputs.Filename(0))
		script := filepath.Join(workdir, "fail.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 7\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if code := runSupervisor(t, workdir, script, outputFile); code != 7 {
			t.Fatalf("expected exit 7, got %d", code)
		}
	})

	// A non-root image can't create anything in a workspace owned by the runner,
	// which used to leave nuon_output uninstalled and outputs silently empty.
	//
	// The workspace is made unwritable two ways because tests run as root in CI,
	// where the mode alone is not enforced: a regular file sits where the old
	// helper directory would go, so creating it fails with ENOTDIR for any uid.
	t.Run("collects nuon_output without writing the workspace", func(t *testing.T) {
		workdir := t.TempDir()
		outputFile := filepath.Join(workdir, outputs.Filename(0))
		script := filepath.Join(workdir, "step.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\nnuon_output hello world\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(outputFile, nil, 0o666); err != nil {
			t.Fatal(err)
		}

		supPath, err := Write(workdir)
		if err != nil {
			t.Fatalf("write supervisor: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workdir, ".nuon"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(workdir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(workdir, 0o755) })

		cmd := exec.CommandContext(context.Background(), "/bin/sh", supPath, "--script", script, "--workdir", workdir)
		cmd.Env = append(os.Environ(),
			RootEnvVar+"="+workdir,
			OutputFilepathEnvVar+"="+outputFile,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("supervisor must not depend on writing the workspace: %v", err)
		}

		out, err := outputs.ParseFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}
		if out["hello"] != "world" {
			t.Fatalf("expected hello=world, got %v", out["hello"])
		}
	})
}
