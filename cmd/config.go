package cmd

import (
	"io/ioutil"
    "path"
    "os"
	"github.com/pelletier/go-toml"
)

type config struct {
	Path string
}

func (config *config) pathDB() string {
	return path.Join(config.Path, "db")
}

func openConfig() (config, error) {
	var config config

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
