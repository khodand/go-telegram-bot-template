package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"

	"go-telegram-bot-template/pkg/telegram"
)

type Config struct {
	General  General
	Telegram telegram.Config
}

type General struct {
	Debug bool
}

func Init() (Config, error) {
	const (
		configPathEnv = "CONFIG_PATH"
	)
	var (
		configData []byte
		err        error
		filePath   string
	)

	if filePath = os.Getenv(configPathEnv); filePath == "" {
		filePath = "config/secret.yaml"
	}

	if configData, err = os.ReadFile(filepath.Clean(filePath)); err != nil {
		return Config{}, fmt.Errorf("read file: %w", err)
	}

	expandedData := os.ExpandEnv(string(configData))

	var cfg Config
	if err = yaml.UnmarshalStrict([]byte(expandedData), &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshall yaml: %w", err)
	}
	if err = envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("evconfig: %w", err)
	}

	return cfg, nil
}
