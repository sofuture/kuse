package common

import (
	"os"
	"path/filepath"
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

	if s.CurrentName() != "development" {
		t.Fatalf("current = %q, want development", s.CurrentName())
	}
	names := s.TargetNames()
	if len(names) != 2 || names[0] != "development" || names[1] != "production" {
		t.Fatalf("targets = %v, want [development production]", names)
	}

	if err := s.SetTarget("production"); err != nil {
		t.Fatal(err)
	}

	s2, err := LoadState(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if s2.CurrentName() != "production" {
		t.Fatalf("current after switch = %q, want production", s2.CurrentName())
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
	if err := s.SetTarget("missing"); err == nil {
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
	if s.CurrentName() != "cluster" {
		t.Fatalf("current = %q, want cluster", s.CurrentName())
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
	if err := s.SetTarget("dev"); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(kubeconfig) {
		t.Fatal("expected kubeconfig symlink to be created")
	}
}
