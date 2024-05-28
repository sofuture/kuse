package common

import (
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

func InitConfig() (*Config, error) {
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

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No kuse configuration found, no sweat, I'll create one with defaults at", cfgLocation)
			err := viper.WriteConfigAs(cfgLocation)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else {
			return nil, err
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
