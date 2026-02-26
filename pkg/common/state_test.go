package common

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTargetsExcludesYamlNamedDirectories(t *testing.T) {
	sourcesDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourcesDir, "dev.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create yaml file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "notes.txt"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("failed to create non-yaml file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(sourcesDir, "bad.yaml"), 0o755); err != nil {
		t.Fatalf("failed to create yaml-named directory: %v", err)
	}

	state := &State{config: &Config{Sources: sourcesDir}}
	if err := state.loadTargets(); err != nil {
		t.Fatalf("loadTargets returned error: %v", err)
	}

	if len(state.targets) != 1 {
		t.Fatalf("expected 1 target, got %d (%v)", len(state.targets), state.targets)
	}
	if state.targets[0].Name != "dev" {
		t.Fatalf("expected dev target, got %q", state.targets[0].Name)
	}
}

func TestLoadCurrentMissingKubeconfig(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), ".kube", "config")
	state := &State{config: &Config{Kubeconfig: missingPath}}

	err := state.loadCurrent()
	if err == nil {
		t.Fatalf("expected error when kubeconfig is missing")
	}
	if err.Error() != "kubeconfig does not exist" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadCurrentNonSymlink(t *testing.T) {
	tempDir := t.TempDir()
	kubeconfigPath := filepath.Join(tempDir, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		t.Fatalf("failed to create kube directory: %v", err)
	}
	if err := os.WriteFile(kubeconfigPath, []byte("plain"), 0o600); err != nil {
		t.Fatalf("failed to create kubeconfig file: %v", err)
	}

	state := &State{config: &Config{Kubeconfig: kubeconfigPath}}
	err := state.loadCurrent()
	if err == nil {
		t.Fatalf("expected error when kubeconfig is not a symlink")
	}
	if err.Error() != "kubeconfig is not a symlink" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadCurrentSymlink(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "kubeconfigs", "dev.yaml")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	kubeconfigPath := filepath.Join(tempDir, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		t.Fatalf("failed to create kube directory: %v", err)
	}
	if err := os.Symlink(targetPath, kubeconfigPath); err != nil {
		t.Fatalf("failed to create kubeconfig symlink: %v", err)
	}

	state := &State{config: &Config{Kubeconfig: kubeconfigPath}}
	if err := state.loadCurrent(); err != nil {
		t.Fatalf("loadCurrent returned error: %v", err)
	}

	if state.current.Name != "dev" {
		t.Fatalf("expected current target name dev, got %q", state.current.Name)
	}
}

func TestSetTargetRejectsInvalidTarget(t *testing.T) {
	state := &State{
		config:  &Config{Kubeconfig: filepath.Join(t.TempDir(), ".kube", "config")},
		targets: []Link{{Name: "dev", File: "/tmp/dev.yaml"}},
	}

	err := state.SetTarget("prod")
	if err == nil {
		t.Fatalf("expected error for invalid target")
	}
	if err.Error() != "invalid target: prod" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetTargetCreatesSymlinkForValidTarget(t *testing.T) {
	tempDir := t.TempDir()
	kubeconfigPath := filepath.Join(tempDir, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		t.Fatalf("failed to create kube directory: %v", err)
	}

	targetPath := filepath.Join(tempDir, "sources", "dev.yaml")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	state := &State{
		config:  &Config{Kubeconfig: kubeconfigPath},
		targets: []Link{{Name: "dev", File: targetPath}},
	}

	if err := state.SetTarget("dev"); err != nil {
		t.Fatalf("SetTarget returned error: %v", err)
	}

	resolvedTarget, err := os.Readlink(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed reading resulting symlink: %v", err)
	}
	if resolvedTarget != targetPath {
		t.Fatalf("expected symlink target %q, got %q", targetPath, resolvedTarget)
	}
}

func TestLoadStateReturnsWarningWhenCurrentMissing(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "kubeconfigs")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("failed to create sources directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "dev.yaml"), []byte("apiVersion: v1\n"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	state, err := LoadState(&Config{
		Kubeconfig: filepath.Join(tempDir, ".kube", "config"),
		Sources:    sourcesDir,
	})
	if err == nil {
		t.Fatalf("expected warning error from LoadState")
	}

	var warning *StateWarning
	if !errors.As(err, &warning) {
		t.Fatalf("expected StateWarning, got %T", err)
	}
	if state.current.Name != "~none~" {
		t.Fatalf("expected fallback current target name, got %q", state.current.Name)
	}
	if len(state.targets) != 1 || state.targets[0].Name != "dev" {
		t.Fatalf("expected targets to load despite warning, got %v", state.targets)
	}
}
