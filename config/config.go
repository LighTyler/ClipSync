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

// createDefaultConfig создаёт файл конфигурации с дефолтными значениями
func initConfig(path string) error {
	// Создаём директорию, если её нет
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("config - initConfig - MkdirAll %s: %w", dir, err)
	}

	// Дефолтный конфиг
	defaultCfg := Config{
		Listen: ":9090",
		Peers:  []string{},
	}

	// Сериализуем с форматированием
	data, err := json.MarshalIndent(defaultCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config - initConfig - MarshalIndent: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("config - initConfig - WriteFile %s: %w", path, err)
	}

	return nil
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
		// Если файл не найден — создаём дефолтный
		if os.IsNotExist(err) {
			if createErr := initConfig(path); createErr != nil {
				return nil, fmt.Errorf("config - LoadConfig - failed to create default config: %w", createErr)
			}
			// Читаем только что созданный файл
			data, err = os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("config - LoadConfig - ReadFile %s after creation: %w", path, err)
			}
		} else {
			return nil, fmt.Errorf("config - LoadConfig - ReadFile %s: %w", path, err)
		}
	}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config - LoadConfig - Unmarshal: %w", err)
	}
	return cfg, nil
}
