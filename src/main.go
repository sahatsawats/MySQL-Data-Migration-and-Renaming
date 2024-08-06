package main

import (
	"MDMR/src/conf"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func makeDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)

		if err != nil {
			return err
		}
	}

	return nil
}

func readingConfigurationFile() *mdmr_config.Configurations {
	// get the current execute directory
	baseDir, err := os.Executable()

	if err != nil {
		log.Fatal(err)
	}

	var conf mdmr_config.Configurations
	configFile := filepath.Join(filepath.Dir(baseDir), "conf", "config.yaml")

	readConf, err := os.ReadFile(configFile)

	if err != nil {
		log.Fatalf("Failred to read configuration file: %v", err)
	}

	err = yaml.Unmarshal(readConf, &conf)

	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	return &conf
}

func main() {
	fmt.Println("Start reading configuration file...")
	
}