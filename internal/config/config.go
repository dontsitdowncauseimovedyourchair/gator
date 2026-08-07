package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't read %v: %w", configFilePath, err)
	}

	var out Config
	err = json.Unmarshal(data, &out)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't read json file: %w", err)
	}
	return out, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	err := write(*c)
	if err != nil {
		return fmt.Errorf("couldn't set user: %w", err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("couldn't retrieve User Home Directory: %w", err)
	}
	return home + "/" + configFileName, nil
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("couldn't encode config: %w", err)
	}
	err = os.WriteFile(configFilePath, encoded, 0777)
	if err != nil {
		return fmt.Errorf("flopped writing data to config file: %w", err)
	}
	return nil
}
