package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	APIKeys map[string]string `json:"api_keys"`
}

func Init_config() *Config {
	// Read the JSON file
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
	}

	// Unmarshal the JSON data into a Config struct
	var configs Config
	err = json.Unmarshal(data, &configs)
	if err != nil {
		log.Fatal(err)
	}
	return &configs

}
