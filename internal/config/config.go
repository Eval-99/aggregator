package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	err := write(c)
	if err != nil {
		return err
	}

	return nil
}

func Read() (Config, error) {
	homeDir, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	filePath := homeDir + "/" + configFileName

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var fileStruct Config
	if err := json.Unmarshal(data, &fileStruct); err != nil {
		return Config{}, err
	}

	return fileStruct, nil
}

func write(cfg *Config) error {
	fileJson, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	homeDir, err := getConfigFilePath()
	if err != nil {
		return err
	}

	filePath := homeDir + "/" + configFileName

	err = os.WriteFile(filePath, fileJson, 0o666)
	if err != nil {
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}
