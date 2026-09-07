package common

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

const (
	configFileLocation = "kuse/kuseconfig.yaml"

	defaultKubeconfig = "~/.kube/config"
	defaultSources    = "~/kubeconfigs"
)

// fileConfig is the on-disk YAML shape (kept stable for existing installs).
type fileConfig struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Sources    string `yaml:"sources"`
}

// Config holds resolved filesystem paths used by kuse.
type Config struct {
	Kubeconfig string
	Sources    string
}

// InitConfig loads (or creates) the kuse configuration, applying optional CLI overrides.
func InitConfig(kubeconfig string, sources string) (*Config, error) {
	cfgLocation, err := xdg.ConfigFile(configFileLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to locate %s: %w", configFileLocation, err)
	}

	raw := fileConfig{
		Kubeconfig: defaultKubeconfig,
		Sources:    defaultSources,
	}

	_, statErr := os.Stat(cfgLocation)
	configExists := statErr == nil
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("stat config: %w", statErr)
	}

	if configExists {
		loaded, err := readFileConfig(cfgLocation)
		if err != nil {
			return nil, err
		}
		if loaded.Kubeconfig != "" {
			raw.Kubeconfig = loaded.Kubeconfig
		}
		if loaded.Sources != "" {
			raw.Sources = loaded.Sources
		}
	}

	if kubeconfig != "" {
		raw.Kubeconfig = kubeconfig
	}
	if sources != "" {
		raw.Sources = sources
	}

	expandedKubeconfig, err := resolvePath(raw.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("expand kubeconfig path: %w", err)
	}
	expandedSources, err := resolvePath(raw.Sources)
	if err != nil {
		return nil, fmt.Errorf("expand sources path: %w", err)
	}

	creatingDefaults := !configExists && kubeconfig == "" && sources == ""

	// Ensure sources is usable before persisting anything, so a typo'd --sources
	// does not leave a broken config behind.
	if creatingDefaults {
		if err := os.MkdirAll(expandedSources, 0o755); err != nil {
			return nil, fmt.Errorf("create sources directory: %w", err)
		}
	} else if err := requireSourcesDir(expandedSources); err != nil {
		return nil, err
	}

	shouldWrite := !configExists || kubeconfig != "" || sources != ""
	if shouldWrite {
		if creatingDefaults {
			fmt.Fprintf(os.Stderr, "No kuse configuration found; creating defaults at %s\n", cfgLocation)
		}
		// Persist resolved absolute paths so relative CLI values do not change
		// meaning when the working directory changes.
		if err := writeFileConfig(cfgLocation, fileConfig{
			Kubeconfig: expandedKubeconfig,
			Sources:    expandedSources,
		}); err != nil {
			return nil, err
		}
	}

	return &Config{
		Kubeconfig: expandedKubeconfig,
		Sources:    expandedSources,
	}, nil
}

func requireSourcesDir(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("sources directory: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("sources is not a directory: %s", path)
	}
	return nil
}

func readFileConfig(path string) (fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, fmt.Errorf("read config: %w", err)
	}
	var cfg fileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func writeFileConfig(path string, cfg fileConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
