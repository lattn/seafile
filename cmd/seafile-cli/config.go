package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

func (c *Config) Save() error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".config", "seafile-cli.json"), b, 0644)
}

func LoadConfig() (Config, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	dir = filepath.Join(dir, ".config", "seafile-cli.json")
	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, errors.New("config initialization is required")
		}
		return Config{}, err
	}
	b, err := os.ReadFile(dir)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	return cfg, err
}
