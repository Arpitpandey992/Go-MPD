package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TODO: make this config file to be outside the DB-interface service (somewhere in .config maybe)
const CONFIG_FILE_PATH = "/home/arpit/Programming/Python/Audio-DB-Interface/config.yml"

type MeiliSearchConfig struct {
	Host            string   `yaml:"host"`
	Port            int      `yaml:"port"`
	DbPath          string   `yaml:"db_path"`
	IndexName       string   `yaml:"index_name"`
	IndexPrimaryKey string   `yaml:"index_primary_key"`
	StartupArgs     []string `yaml:"startup_args"`
}

type DatabaseConfig struct {
	Meilisearch MeiliSearchConfig `yaml:"meilisearch"`
}

type AudioConfig struct {
	ScanDirectories []string `yaml:"scan_directories"`
	ScanFormats     []string `yaml:"scan_formats"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Audio    AudioConfig    `yaml:"audio"`
}

func GetBaseConfiguration() (*Config, error) {
	fileContent, err := os.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	return &config, nil
}
