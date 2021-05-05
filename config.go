package domain

import (
	"io/ioutil"
    "path"
    "os"
	"github.com/pelletier/go-toml"
)

type Config struct {
	Path string
}

func OpenConfig() (Config, error) {
	var config Config

    userConfigPath, err := os.UserConfigDir()
	if err != nil {
		return config, err
	}

    configPath := path.Join(userConfigPath, "domain", "config.toml")

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}

    homePath, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}

    config.Path = path.Join(homePath, config.Path)

	return config, nil
}
