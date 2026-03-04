package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	var cfg Config

	configPath, _ := getConfigFilePath()

	file, err := os.Open(configPath)
	if err != nil {
		return cfg, fmt.Errorf("could not open config file: %w\n", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("could not decode config JSON: %w\n", err)
	}

	// fmt.Printf("%s", cfg.DbURL)
	// fmt.Printf("%s", cfg.CurrentUserName)
	return cfg, nil
}

// Writes config struct to the JSON file after setting username
func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username

	configPath, _ := getConfigFilePath()

	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open config file for writing: %w\n", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // pretty-print JSON

	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("could not encode config to JSON: %w\n", err)
	}

	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return homeDir, fmt.Errorf("could not determine home directory: %w\n", err)
	}

	configPath := filepath.Join(homeDir, configFileName)

	return configPath, nil
}
