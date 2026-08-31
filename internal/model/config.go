package model

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// Название приложения.
	NAME = "valid-access-back"

	// Имя переменной окружения определяющая режим работы приложения.
	MODE = "MODE"

	// Режим разработки.
	MODE_DEV = "dev"

	// Путь в директории конфигурации в производстве.
	CONFIG_DIRECTORY_PROD = "/etc"

	// Путь в директории конфигурации в разработке.
	CONFIG_DIRECTORY_DEV = "./config"

	// Имя файла конфигурации в производстве.
	CONFIG_FILENAME_PROD = "prod.yaml"

	// Имя файла конфигурации в разработке.
	CONFIG_FILENAME_DEV = "dev.yaml"

	ACCESS_KEY = "ACCESS_KEY"
)

// Conf Глобальный экземпляр конфигурации.
var Conf *Config

// Config Конфигурация.
type Config struct {
	Server    *ServerConfig `yaml:"server"`
	AccessKey string
}

// ServerConfig Конфигурация сервера.
type ServerConfig struct {
	Port uint16 `yaml:"port"`
}

// NewConfig Конфигурация.
func NewConfig() (*Config, error) {
	Log.Info.Info("configuration loading")
	f, err := os.ReadFile(filepath.Join(configDir(), configFilename()))
	if err != nil {
		return nil, err
	}
	ak, ok := os.LookupEnv(ACCESS_KEY)
	if !ok {
		return nil, fmt.Errorf("env %s not set", ACCESS_KEY)
	}

	c := &Config{
		Server:    &ServerConfig{},
		AccessKey: ak,
	}
	if err := yaml.Unmarshal(f, c); err != nil {
		return nil, err
	}
	return c, nil
}

// configDir Директоия конфигурации.
func configDir() string {
	m, o := os.LookupEnv(MODE)
	if o && m == MODE_DEV {
		return CONFIG_DIRECTORY_DEV
	}
	return filepath.Join(CONFIG_DIRECTORY_PROD, NAME)
}

// configFilename Имя файла конфигурации.
func configFilename() string {
	m, o := os.LookupEnv(MODE)
	if o && m == MODE_DEV {
		return CONFIG_FILENAME_DEV
	}
	return CONFIG_FILENAME_PROD
}
