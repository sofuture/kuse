package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func TestInitConfigCreatesDefaults(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
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
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := InitConfig(kube, sources)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Kubeconfig != kube || cfg.Sources != sources {
		t.Fatalf("unexpected config: %+v", cfg)
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
	if err := os.MkdirAll(customSources, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitConfig(customKube, customSources); err != nil {
		t.Fatal(err)
	}

	newSources := filepath.Join(home, "other-sources")
	if err := os.MkdirAll(newSources, 0o755); err != nil {
		t.Fatal(err)
	}
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

func TestInitConfigRejectsUserSpecificHome(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := InitConfig("", "~someuser/cfgs")
	if err == nil {
		t.Fatal("expected error for ~user path")
	}
	if !strings.Contains(err.Error(), "cannot expand user-specific home dir") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bad value must not be persisted, and no literal ~someuser dir created in CWD.
	cfgPath := filepath.Join(configHome, "kuse", "kuseconfig.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		data, _ := os.ReadFile(cfgPath)
		if strings.Contains(string(data), "~someuser") {
			t.Fatalf("persisted bad sources value: %s", data)
		}
	}
}

func TestInitConfigDoesNotCreateOverrideSources(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	missing := filepath.Join(home, "does-not-exist-yet")
	cfg, err := InitConfig("", missing)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sources != missing {
		t.Fatalf("Sources = %q, want %q", cfg.Sources, missing)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("expected override sources not to be auto-created, err=%v", err)
	}
}
