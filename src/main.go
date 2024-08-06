package main

import (
	"flag"
	"os"

	"MDMR/src/config"
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

