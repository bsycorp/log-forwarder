package main

import (
	"io/ioutil"
	"log"
	"os"
)

func ReadStateFile(stateFile string) (string, error) {
	b, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("State file not found")
			return "", nil
		}
		log.Println("Error reading state file:", err)
		return "", err
	}
	return string(b), nil
}

func WriteStateFile(stateFile string, cursor string) error {
	err := ioutil.WriteFile(stateFile, []byte(cursor), os.FileMode(0600))
	if err != nil {
		log.Println("Error writing state file:", err)
	}
	return err
}
