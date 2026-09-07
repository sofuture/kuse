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
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("expected config file at %s: %v", cfgPath, err)
	}
	if !strings.Contains(string(data), wantKube) || !strings.Contains(string(data), wantSources) {
		t.Fatalf("expected absolute paths persisted, got: %s", data)
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
	_, err := InitConfig("", missing)
	if err == nil {
		t.Fatal("expected error for missing --sources override")
	}
	if !strings.Contains(err.Error(), "sources directory") {
		t.Fatalf("unexpected error: %v", err)
	}

	cfgPath := filepath.Join(configHome, "kuse", "kuseconfig.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		t.Fatal("expected missing --sources not to persist a config file")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("expected override sources not to be auto-created, err=%v", err)
	}
}

func TestInitConfigPersistsAbsolutePaths(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	work := t.TempDir()
	t.Chdir(work)
	sources := filepath.Join(work, "cfgs")
	if err := os.MkdirAll(sources, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sources, "from-a.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := InitConfig("", "./cfgs")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sources != sources {
		t.Fatalf("Sources = %q, want %q", cfg.Sources, sources)
	}

	cfgPath := filepath.Join(configHome, "kuse", "kuseconfig.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "sources: ./cfgs") {
		t.Fatalf("persisted relative sources: %s", data)
	}
	if !strings.Contains(string(data), sources) {
		t.Fatalf("expected absolute sources in config, got: %s", data)
	}

	// A later invocation from another cwd should still see the same sources.
	other := t.TempDir()
	t.Chdir(other)
	cfg2, err := InitConfig("", "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.Sources != sources {
		t.Fatalf("Sources after cwd change = %q, want %q", cfg2.Sources, sources)
	}
}

func TestInitConfigTypoDoesNotPersistOverExisting(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	home := t.TempDir()
	t.Setenv("HOME", home)

	goodSources := filepath.Join(home, "good-sources")
	if err := os.MkdirAll(goodSources, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitConfig(filepath.Join(home, "kube"), goodSources); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(configHome, "kuse", "kuseconfig.yaml")
	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = InitConfig("", filepath.Join(home, "missing-typo"))
	if err == nil {
		t.Fatal("expected error for typo'd sources")
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("config changed after failed override:\nbefore=%s\nafter=%s", before, after)
	}
}
