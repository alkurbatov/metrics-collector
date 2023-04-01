package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

// loadFromFile overwrites the provided config with data from configuration file.
func loadFromFile[T any](src entity.FilePath, dst *T) error {
	data, err := os.ReadFile(src.String())
	if err != nil {
		return fmt.Errorf("config - loadFromFile - os.ReadFile: %w", err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("config - loadFromFile - json.Unmarshal: %w", err)
	}

	return nil
}
