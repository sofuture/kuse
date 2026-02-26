package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

func configureTestEnvironment(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))
	xdg.Reload()

	return homeDir
}

func readConfigFile(t *testing.T, configPath string) *viper.Viper {
	t.Helper()

	config := viper.New()
	config.SetConfigFile(configPath)
	config.SetConfigType("yaml")
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	return config
}

func TestInitConfigBootstrapsDefaults(t *testing.T) {
	homeDir := configureTestEnvironment(t)

	config, err := InitConfig("", "")
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}

	expectedKubeconfig := filepath.Join(homeDir, ".kube", "config")
	if config.Kubeconfig != expectedKubeconfig {
		t.Fatalf("expected kubeconfig %q, got %q", expectedKubeconfig, config.Kubeconfig)
	}

	expectedSources := filepath.Join(homeDir, "kubeconfigs")
	if config.Sources != expectedSources {
		t.Fatalf("expected sources %q, got %q", expectedSources, config.Sources)
	}

	configPath := filepath.Join(homeDir, ".config", "kuse", "kuseconfig.yaml")
	fileConfig := readConfigFile(t, configPath)

	if fileConfig.GetString(keyKubeconfig) != defaultKubeconfig {
		t.Fatalf("expected persisted kubeconfig %q, got %q", defaultKubeconfig, fileConfig.GetString(keyKubeconfig))
	}
	if fileConfig.GetString(keySources) != defaultSources {
		t.Fatalf("expected persisted sources %q, got %q", defaultSources, fileConfig.GetString(keySources))
	}
}

func TestInitConfigPreservesExistingSourcesWhenOverridingKubeconfig(t *testing.T) {
	homeDir := configureTestEnvironment(t)
	configDir := filepath.Join(homeDir, ".config", "kuse")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	existingSources := filepath.Join(homeDir, "sources-a")
	if err := os.MkdirAll(existingSources, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	configPath := filepath.Join(configDir, "kuseconfig.yaml")
	if err := os.WriteFile(configPath, []byte(
		"kubeconfig: "+filepath.Join(homeDir, ".kube", "old-config")+"\n"+
			"sources: "+existingSources+"\n",
	), 0o600); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	newKubeconfig := filepath.Join(homeDir, ".kube", "new-config")
	config, err := InitConfig(newKubeconfig, "")
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}

	if config.Kubeconfig != newKubeconfig {
		t.Fatalf("expected kubeconfig %q, got %q", newKubeconfig, config.Kubeconfig)
	}
	if config.Sources != existingSources {
		t.Fatalf("expected sources %q, got %q", existingSources, config.Sources)
	}

	fileConfig := readConfigFile(t, configPath)
	if fileConfig.GetString(keyKubeconfig) != newKubeconfig {
		t.Fatalf("expected persisted kubeconfig %q, got %q", newKubeconfig, fileConfig.GetString(keyKubeconfig))
	}
	if fileConfig.GetString(keySources) != existingSources {
		t.Fatalf("expected persisted sources %q, got %q", existingSources, fileConfig.GetString(keySources))
	}
}

func TestInitConfigPreservesExistingKubeconfigWhenOverridingSources(t *testing.T) {
	homeDir := configureTestEnvironment(t)
	configDir := filepath.Join(homeDir, ".config", "kuse")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	existingKubeconfig := filepath.Join(homeDir, ".kube", "current-config")
	configPath := filepath.Join(configDir, "kuseconfig.yaml")
	if err := os.WriteFile(configPath, []byte(
		"kubeconfig: "+existingKubeconfig+"\n"+
			"sources: "+filepath.Join(homeDir, "old-sources")+"\n",
	), 0o600); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	newSources := filepath.Join(homeDir, "new-sources")
	if err := os.MkdirAll(newSources, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	config, err := InitConfig("", newSources)
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}

	if config.Kubeconfig != existingKubeconfig {
		t.Fatalf("expected kubeconfig %q, got %q", existingKubeconfig, config.Kubeconfig)
	}
	if config.Sources != newSources {
		t.Fatalf("expected sources %q, got %q", newSources, config.Sources)
	}

	fileConfig := readConfigFile(t, configPath)
	if fileConfig.GetString(keyKubeconfig) != existingKubeconfig {
		t.Fatalf("expected persisted kubeconfig %q, got %q", existingKubeconfig, fileConfig.GetString(keyKubeconfig))
	}
	if fileConfig.GetString(keySources) != newSources {
		t.Fatalf("expected persisted sources %q, got %q", newSources, fileConfig.GetString(keySources))
	}
}

func TestInitConfigExpandsTildePaths(t *testing.T) {
	homeDir := configureTestEnvironment(t)
	configDir := filepath.Join(homeDir, ".config", "kuse")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "kuseconfig.yaml")
	contents := strings.Join([]string{
		"kubeconfig: ~/.kube/custom-config",
		"sources: ~/custom-sources",
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	config, err := InitConfig("", "")
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}

	if strings.HasPrefix(config.Kubeconfig, "~") {
		t.Fatalf("expected kubeconfig to be expanded, got %q", config.Kubeconfig)
	}
	if strings.HasPrefix(config.Sources, "~") {
		t.Fatalf("expected sources to be expanded, got %q", config.Sources)
	}
	if !strings.HasSuffix(config.Kubeconfig, filepath.Join(".kube", "custom-config")) {
		t.Fatalf("expected kubeconfig to end with .kube/custom-config, got %q", config.Kubeconfig)
	}
	if !strings.HasSuffix(config.Sources, filepath.Join("custom-sources")) {
		t.Fatalf("expected sources to end with custom-sources, got %q", config.Sources)
	}
}
