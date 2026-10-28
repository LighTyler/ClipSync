package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultConfigPath = "./config/config.json"

type Config struct {
	Peers  []string `json:"peers"`
	Listen string   `json:"listen"`
}

// GetDefaultConfigPath возвращает правильный путь к конфигу в зависимости от ОС
func GetDefaultConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Если по какой-то причине не удалось получить системную папку (редкость),
		// fallback на текущую директорию
		return "config.json"
	}

	// Создаем подпапку с именем нашего приложения
	appDir := filepath.Join(configDir, "ClipSync")
	return filepath.Join(appDir, "config.json")
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config - NewConfig - ReadFile: %w", err)
	}
	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config - NewConfig - Unmarshal: %w", err)
	}
	return cfg, nil
}
