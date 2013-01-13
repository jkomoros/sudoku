package main

import (
	"encoding/json"
	"log"
	"os"
)

const _DB_CONFIG_FILENAME = "db_config.SECRET.json"

type dbConfig struct {
	Url, Username, Password string
}

func main() {
	file, err := os.Open(_DB_CONFIG_FILENAME)
	if err != nil {
		log.Fatal("Could not find the config file at ", _DB_CONFIG_FILENAME, ". You should copy the SAMPLE one to that filename and configure.")
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var config dbConfig
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("There was an error parsing JSON from the config file: ", err)
		os.Exit(1)
	}
	log.Println(config)

}
