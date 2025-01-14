package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() *Config {
	cfgFilePath, err := getConfigFilePath()
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return nil
	}

	cfgFile, err := os.ReadFile(cfgFilePath)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return nil
	}

	c := Config{}

	err = json.Unmarshal(cfgFile, &c)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return nil
	}

	return &c
}

func (cfg *Config) SetUser() {
	cfg.CurrentUserName = "jose.nieto"

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return
	}

	cfgFilePath, err := getConfigFilePath()
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return
	}

	err = os.WriteFile(cfgFilePath, jsonData, 0644)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		return
	}
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	return filepath.Join(homeDir, configFileName), err
}

func write(cfg Config) error {

	return nil
}
