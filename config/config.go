package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	MAC   string `toml:"mac"`
	Color bool   `toml:"color"`
}

func Default() *Config {
	return &Config{
		Color: true,
	}
}

func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.Getenv("HOME")
	}
	return filepath.Join(dir, "aranet4-go", "config.toml")
}

func Load() (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
