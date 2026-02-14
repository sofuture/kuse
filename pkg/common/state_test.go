package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetTarget_ReplacesBrokenKubeconfigSymlink(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll sources returned error: %v", err)
	}

	targetFile := filepath.Join(sourcesDir, "dev.yaml")
	if err := os.WriteFile(targetFile, []byte("apiVersion: v1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile target returned error: %v", err)
	}

	kubeconfigPath := filepath.Join(tempDir, ".kube", "config")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		t.Fatalf("MkdirAll kubeconfig parent returned error: %v", err)
	}

	brokenTarget := filepath.Join(sourcesDir, "missing.yaml")
	if err := os.Symlink(brokenTarget, kubeconfigPath); err != nil {
		t.Fatalf("Symlink returned error: %v", err)
	}

	state, err := LoadState(&Config{
		Kubeconfig: kubeconfigPath,
		Sources:    sourcesDir,
	})
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}

	if state.current.Name != "missing" {
		t.Fatalf("current.Name = %q, want %q", state.current.Name, "missing")
	}

	if err := state.SetTarget("dev"); err != nil {
		t.Fatalf("SetTarget returned error: %v", err)
	}

	resolvedLink, err := os.Readlink(kubeconfigPath)
	if err != nil {
		t.Fatalf("Readlink returned error: %v", err)
	}
	if resolvedLink != targetFile {
		t.Fatalf("kubeconfig symlink target = %q, want %q", resolvedLink, targetFile)
	}
}
