package common

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStateAndSetTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeDir := filepath.Join(root, ".kube")
	kubeconfig := filepath.Join(kubeDir, "config")

	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(kubeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	dev := filepath.Join(sources, "development.yaml")
	prod := filepath.Join(sources, "production.yaml")
	if err := os.WriteFile(dev, []byte("apiVersion: v1\nkind: Config\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prod, []byte("apiVersion: v1\nkind: Config\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Non-yaml files should be ignored.
	if err := os.WriteFile(filepath.Join(sources, "readme.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(dev, kubeconfig); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Kubeconfig: kubeconfig, Sources: sources}
	s, err := LoadState(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if s.current.Name != "development" {
		t.Fatalf("current = %q, want development", s.current.Name)
	}
	names := s.targetNames()
	if len(names) != 2 || names[0] != "development" || names[1] != "production" {
		t.Fatalf("targets = %v, want [development production]", names)
	}

	if err := s.SetTarget("production", false); err != nil {
		t.Fatal(err)
	}

	s2, err := LoadState(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if s2.current.Name != "production" {
		t.Fatalf("current after switch = %q, want production", s2.current.Name)
	}

	link, err := os.Readlink(kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	if link != prod {
		t.Fatalf("symlink target = %q, want %q", link, prod)
	}
}

func TestSetTargetInvalid(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeconfig := filepath.Join(root, "config")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sources, "only.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetTarget("missing", false); err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestLoadCurrentRelativeSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeDir := filepath.Join(root, ".kube")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(kubeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	abs := filepath.Join(sources, "cluster.yaml")
	if err := os.WriteFile(abs, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Place a relative symlink from .kube/config -> ../kubeconfigs/cluster.yaml
	kubeconfig := filepath.Join(kubeDir, "config")
	rel, err := filepath.Rel(kubeDir, abs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rel, kubeconfig); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if s.current.Name != "cluster" {
		t.Fatalf("current = %q, want cluster", s.current.Name)
	}
}

func TestSetTargetCreatesParentDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeconfig := filepath.Join(root, "nested", "kube", "config")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sources, "dev.yaml")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetTarget("dev", false); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(kubeconfig) {
		t.Fatal("expected kubeconfig symlink to be created")
	}
}

func TestSetTargetForceOverwritesRegularFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeconfig := filepath.Join(root, "config")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sources, "dev.yaml")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kubeconfig, []byte("regular file"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetTarget("dev", true); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(kubeconfig) {
		t.Fatal("expected force overwrite to create symlink")
	}
	link, err := os.Readlink(kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	if link != target {
		t.Fatalf("symlink target = %q, want %q", link, target)
	}
}

func TestLoadTargetsSkipsHiddenFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sources, "visible.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sources, ".hidden.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: filepath.Join(root, "missing"), Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	names := s.targetNames()
	if len(names) != 1 || names[0] != "visible" {
		t.Fatalf("targets = %v, want [visible]", names)
	}
}

func TestPrintStatusCommandWritesCurrentError(t *testing.T) {
	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeconfig := filepath.Join(root, "config")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kubeconfig, []byte("regular"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if s.currentErr == nil {
		t.Fatal("expected currentErr for non-symlink kubeconfig")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = old
		_ = w.Close()
		_ = r.Close()
	})

	s.PrintStatusCommand()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "kubeconfig is not a symlink") {
		t.Fatalf("stderr = %q, want symlink error", out)
	}
	if !strings.Contains(out, "--force") {
		t.Fatalf("stderr = %q, want --force hint", out)
	}
}

func TestSetTargetReplacesDanglingSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := filepath.Join(root, "kubeconfigs")
	kubeDir := filepath.Join(root, ".kube")
	kubeconfig := filepath.Join(kubeDir, "config")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(kubeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sources, "dev.yaml")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(sources, "missing.yaml"), kubeconfig); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState(&Config{Kubeconfig: kubeconfig, Sources: sources})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetTarget("dev", false); err != nil {
		t.Fatal(err)
	}
	link, err := os.Readlink(kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	if link != target {
		t.Fatalf("symlink target = %q, want %q", link, target)
	}
}
