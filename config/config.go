package config

import (
	"encoding/json"
	"github.com/unielon-org/unielon-indexer/utils"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Server     utils.ServerConfig `json:"server"`
	Sqlite     utils.SqliteConfig `json:"sqlite"`
	Chain      utils.ChainConfig  `json:"chain"`
	DebugLevel int                `json:"debug_level"`
}

func LoadConfig(cfg *Config, filep string) {

	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filep != "" {
		configFileName = filep
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
