package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Token string `yaml:"token"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open(filepath.Join("d:\\", "projects", "golang", "tg-program", "client-tg", "config", "token.yaml"))
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге YAML: %v", err)
	}

	return &config, nil
}
