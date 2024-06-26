package common

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"path"
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

	viper.SetDefault(keyKubeconfig, defaultKubeconfig)
	viper.SetDefault(keySources, defaultSources)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileExtension)
	viper.AddConfigPath(path.Dir(cfgLocation))

	if kubeconfig != "" {
		viper.Set(keyKubeconfig, kubeconfig)
	}

	if sources != "" {
		viper.Set(keySources, sources)
	}

	if kubeconfig != "" || sources != "" {
		err := viper.WriteConfigAs(cfgLocation)
		if err != nil {
			return nil, err
		}
	}

	err = viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			fmt.Println("No kuse configuration found, no sweat, I'll create one with defaults at", cfgLocation)
			err := viper.WriteConfigAs(cfgLocation)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
	}

	expandedKubeconfig, err := homedir.Expand(viper.GetString(keyKubeconfig))
	if err != nil {
		return nil, err
	}

	expandedSources, err := homedir.Expand(viper.GetString(keySources))
	if err != nil {
		return nil, err
	}

	return &Config{
		Kubeconfig: expandedKubeconfig,
		Sources:    expandedSources,
	}, nil
}
