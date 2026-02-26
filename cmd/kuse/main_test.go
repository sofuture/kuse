package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildKuseBinary(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("unable to determine test file location")
	}

	packageDir := filepath.Dir(filename)
	binPath := filepath.Join(t.TempDir(), "kuse")

	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = packageDir
	build.Env = os.Environ()

	output, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build kuse binary: %v\n%s", err, string(output))
	}

	return binPath
}

func testEnvForHome(homeDir string) []string {
	return append(os.Environ(),
		"HOME="+homeDir,
		"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
	)
}

func writeConfig(t *testing.T, homeDir string, kubeconfig string, sources string) {
	t.Helper()

	configDir := filepath.Join(homeDir, ".config", "kuse")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "kuseconfig.yaml")

	content := strings.Join([]string{
		"kubeconfig: " + kubeconfig,
		"sources: " + sources,
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func TestShortModeOutputsOnlyCurrentTokenOnStdout(t *testing.T) {
	binary := buildKuseBinary(t)
	homeDir := t.TempDir()

	sourcesDir := filepath.Join(homeDir, "kubeconfigs")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "dev.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	writeConfig(t, homeDir, kubeconfigPath, sourcesDir)

	cmd := exec.Command(binary, "--short")
	cmd.Env = testEnvForHome(homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("short mode command failed: %v\nstderr: %s", err, stderr.String())
	}

	if stdout.String() != "~none~" {
		t.Fatalf("expected stdout token ~none~, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: kubeconfig does not exist") {
		t.Fatalf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestErrorsArePrintedToStderr(t *testing.T) {
	binary := buildKuseBinary(t)
	homeDir := t.TempDir()

	sourcesDir := filepath.Join(homeDir, "kubeconfigs")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}
	target := filepath.Join(sourcesDir, "dev.yaml")
	if err := os.WriteFile(target, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		t.Fatalf("failed to create kube directory: %v", err)
	}
	if err := os.Symlink(target, kubeconfigPath); err != nil {
		t.Fatalf("failed to create kubeconfig symlink: %v", err)
	}

	writeConfig(t, homeDir, kubeconfigPath, sourcesDir)

	cmd := exec.Command(binary, "prod")
	cmd.Env = testEnvForHome(homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected command to fail for invalid target")
	}

	if stdout.String() != "" {
		t.Fatalf("expected no stdout on error, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "error: invalid target: prod") {
		t.Fatalf("expected invalid target error on stderr, got %q", stderr.String())
	}
}
