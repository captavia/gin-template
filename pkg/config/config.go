package config

import (
	"github.com/pelletier/go-toml/v2"
	"io"
	"os"
)

type Config struct {
	Listen string `toml:"listen"`
}

func (c *Config) Default() *Config {
	c.Listen = "0.0.0.0:8000"
	return c
}

func InitConfig(path string) error {
	var config Config

	f, fileErr := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if fileErr != nil {
		return fileErr
	}
	configString, marshalError := toml.Marshal(config.Default())
	if marshalError != nil {
		return fileErr
	}
	_, writeErr := f.Write(configString)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {
	file, openFileErr := os.Open(path)
	defer file.Close()
	if openFileErr != nil {
		return nil, openFileErr
	}

	data, readErr := io.ReadAll(file)
	if readErr != nil {
		return nil, readErr
	}
	var config Config
	config.Default()

	unmarshalErr := toml.Unmarshal(data, &config)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &config, nil
}
