package models

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Key string
}

func getConfig() Configuration {
	dir := "/go/src/github.com/GeoServer/GeoServer/project/api"
	file, err := os.Open(dir + "/conf.json")
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}
