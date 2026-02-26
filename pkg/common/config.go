package common

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"path/filepath"
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

type Config struct {
	Kubeconfig string
	Sources    string
}

func InitConfig(kubeconfig string, sources string) (*Config, error) {
	cfgLocation, err := xdg.ConfigFile(configFileLocation)
	if err != nil {
		fmt.Println("unable to locate", configFileLocation)
		return nil, err
	}

	config := viper.New()
	config.SetDefault(keyKubeconfig, defaultKubeconfig)
	config.SetDefault(keySources, defaultSources)
	config.SetConfigName(configFileName)
	config.SetConfigType(configFileExtension)
	config.AddConfigPath(filepath.Dir(cfgLocation))

	configBootstrapNeeded := false
	if err := config.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			configBootstrapNeeded = true
		} else {
			return nil, err
		}
	}

	if kubeconfig != "" {
		config.Set(keyKubeconfig, kubeconfig)
	}

	if sources != "" {
		config.Set(keySources, sources)
	}

	if configBootstrapNeeded || kubeconfig != "" || sources != "" {
		if err := config.WriteConfigAs(cfgLocation); err != nil {
			return nil, err
		}
	}

	expandedKubeconfig, err := homedir.Expand(config.GetString(keyKubeconfig))
	if err != nil {
		return nil, err
	}

	expandedSources, err := homedir.Expand(config.GetString(keySources))
	if err != nil {
		return nil, err
	}

	return &Config{
		Kubeconfig: expandedKubeconfig,
		Sources:    expandedSources,
	}, nil
}
