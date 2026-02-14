package common

import (
	"os"
	"path/filepath"
	"testing"
)

func setupConfigTestEnv(t *testing.T) (string, string) {
	t.Helper()

	xdgConfigHome := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	t.Setenv("HOME", homeDir)

	return xdgConfigHome, homeDir
}

func TestInitConfig_PartialUpdatePreservesExistingValues(t *testing.T) {
	_, homeDir := setupConfigTestEnv(t)

	initialKubeconfig := filepath.Join(homeDir, ".kube", "config")
	updatedKubeconfig := filepath.Join(homeDir, ".kube", "config-updated")
	customSources := filepath.Join(t.TempDir(), "kubeconfigs")

	if _, err := InitConfig(initialKubeconfig, customSources); err != nil {
		t.Fatalf("InitConfig (initial write) returned error: %v", err)
	}

	updatedConfig, err := InitConfig(updatedKubeconfig, "")
	if err != nil {
		t.Fatalf("InitConfig (partial update) returned error: %v", err)
	}

	if updatedConfig.Kubeconfig != updatedKubeconfig {
		t.Fatalf("Kubeconfig = %q, want %q", updatedConfig.Kubeconfig, updatedKubeconfig)
	}
	if updatedConfig.Sources != customSources {
		t.Fatalf("Sources = %q, want %q", updatedConfig.Sources, customSources)
	}

	reloadedConfig, err := InitConfig("", "")
	if err != nil {
		t.Fatalf("InitConfig (reload) returned error: %v", err)
	}

	if reloadedConfig.Kubeconfig != updatedKubeconfig {
		t.Fatalf("reloaded Kubeconfig = %q, want %q", reloadedConfig.Kubeconfig, updatedKubeconfig)
	}
	if reloadedConfig.Sources != customSources {
		t.Fatalf("reloaded Sources = %q, want %q", reloadedConfig.Sources, customSources)
	}
}

func TestInitConfig_ReturnsErrorForInvalidConfigFile(t *testing.T) {
	xdgConfigHome, _ := setupConfigTestEnv(t)

	configPath := filepath.Join(xdgConfigHome, filepath.FromSlash(configFileLocation))
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("kubeconfig: [unterminated\n"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, err := InitConfig("", ""); err == nil {
		t.Fatal("InitConfig returned nil error for invalid config")
	}
}
