package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	configFileName      = "kuseconfig"
	configFileExtension = "yaml"
	configFileLocation  = "kuse/" + configFileName + "." + configFileExtension

	defaultKubeconfig = "~/.kube/config"
	defaultSources    = "~/kubeconfigs"

	keyKubeconfig = "kubeconfig"
	keySources    = "sources"
)

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

	v := viper.New()
	v.SetDefault(keyKubeconfig, defaultKubeconfig)
	v.SetDefault(keySources, defaultSources)
	v.SetConfigName(configFileName)
	v.SetConfigType(configFileExtension)
	v.AddConfigPath(filepath.Dir(cfgLocation))

	if kubeconfig != "" {
		v.Set(keyKubeconfig, kubeconfig)
	}
	if sources != "" {
		v.Set(keySources, sources)
	}

	if kubeconfig != "" || sources != "" {
		if err := v.WriteConfigAs(cfgLocation); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			fmt.Fprintf(os.Stderr, "No kuse configuration found; creating defaults at %s\n", cfgLocation)
			if err := v.WriteConfigAs(cfgLocation); err != nil {
				return nil, fmt.Errorf("create config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	expandedKubeconfig, err := expandHome(v.GetString(keyKubeconfig))
	if err != nil {
		return nil, fmt.Errorf("expand kubeconfig path: %w", err)
	}
	expandedSources, err := expandHome(v.GetString(keySources))
	if err != nil {
		return nil, fmt.Errorf("expand sources path: %w", err)
	}

	if err := os.MkdirAll(expandedSources, 0o755); err != nil {
		return nil, fmt.Errorf("create sources directory: %w", err)
	}

	return &Config{
		Kubeconfig: expandedKubeconfig,
		Sources:    expandedSources,
	}, nil
}
