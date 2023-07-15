package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/uussoop/simple_proxy/database"
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

func Init_users() {

	data, err := os.ReadFile("users.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Define the struct to hold the JSON data
	type UserData struct {
		Users []database.User `json:"users"`
	}

	// Unmarshal JSON data into UserData struct
	var userData UserData
	err = json.Unmarshal(data, &userData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Insert users into the database
	err = database.InsertUsers(userData.Users)
	if err != nil {
		fmt.Println("Error inserting users:", err)
		return
	}

}
