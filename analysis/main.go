package main

import (
	"encoding/json"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"log"
	"os"
)

const _DB_CONFIG_FILENAME = "db_config.SECRET.json"

type dbConfig struct {
	Url, Username, Password, DbName, SolvesTable, SolvesID, SolvesTotalTime string
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

	db := mysql.New("tcp", "", config.Url, config.Username, config.Password, config.DbName)

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	rows, _, err := db.Query("select %s, %s from %s limit 100", config.SolvesID, config.SolvesTotalTime, config.SolvesTable)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for _, row := range rows {
		log.Println(row.Str(0), row.Str(1))
	}
}
