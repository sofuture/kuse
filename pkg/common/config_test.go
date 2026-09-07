package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestInitConfigCreatesDefaults(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	// xdg caches ConfigHome at init; override via package var if available.
	// adrg/xdg reads env on first use of ConfigFile when Reload is called.
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := InitConfig("", "")
	if err != nil {
		t.Fatal(err)
	}

	wantKube := filepath.Join(home, ".kube", "config")
	wantSources := filepath.Join(home, "kubeconfigs")
	if cfg.Kubeconfig != wantKube {
		t.Fatalf("Kubeconfig = %q, want %q", cfg.Kubeconfig, wantKube)
	}
	if cfg.Sources != wantSources {
		t.Fatalf("Sources = %q, want %q", cfg.Sources, wantSources)
	}

	cfgPath := filepath.Join(configHome, "kuse", "kuseconfig.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("expected config file at %s: %v", cfgPath, err)
	}
	if st, err := os.Stat(wantSources); err != nil || !st.IsDir() {
		t.Fatalf("expected sources dir created at %s: %v", wantSources, err)
	}
}

func TestInitConfigOverrides(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	kube := filepath.Join(home, "custom-kube")
	sources := filepath.Join(home, "custom-sources")

	cfg, err := InitConfig(kube, sources)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Kubeconfig != kube || cfg.Sources != sources {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if st, err := os.Stat(sources); err != nil || !st.IsDir() {
		t.Fatalf("expected sources dir: %v", err)
	}
}

func TestInitConfigPreservesUnsetFieldsOnPartialOverride(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	customKube := filepath.Join(home, "my-kube")
	customSources := filepath.Join(home, "my-sources")
	if _, err := InitConfig(customKube, customSources); err != nil {
		t.Fatal(err)
	}

	newSources := filepath.Join(home, "other-sources")
	cfg, err := InitConfig("", newSources)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Kubeconfig != customKube {
		t.Fatalf("Kubeconfig = %q, want preserved %q", cfg.Kubeconfig, customKube)
	}
	if cfg.Sources != newSources {
		t.Fatalf("Sources = %q, want %q", cfg.Sources, newSources)
	}
}
