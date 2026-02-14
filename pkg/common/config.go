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

	v := viper.New()
	v.SetDefault(keyKubeconfig, defaultKubeconfig)
	v.SetDefault(keySources, defaultSources)

	v.SetConfigName(configFileName)
	v.SetConfigType(configFileExtension)
	v.AddConfigPath(filepath.Dir(cfgLocation))

	shouldWriteConfig := false
	err = v.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			shouldWriteConfig = true
			fmt.Println("No kuse configuration found, no sweat, I'll create one with defaults at", cfgLocation)
		} else {
			return nil, err
		}
	}

	if kubeconfig != "" {
		v.Set(keyKubeconfig, kubeconfig)
		shouldWriteConfig = true
	}

	if sources != "" {
		v.Set(keySources, sources)
		shouldWriteConfig = true
	}

	if shouldWriteConfig {
		err := v.WriteConfigAs(cfgLocation)
		if err != nil {
			return nil, err
		}
	}

	expandedKubeconfig, err := homedir.Expand(v.GetString(keyKubeconfig))
	if err != nil {
		return nil, err
	}

	expandedSources, err := homedir.Expand(v.GetString(keySources))
	if err != nil {
		return nil, err
	}

	return &Config{
		Kubeconfig: expandedKubeconfig,
		Sources:    expandedSources,
	}, nil
}
