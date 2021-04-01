package main

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func main() {
	dirKeys := "sm2keys"

	// Load the config, generate and save keys
	filePath := "cmd/sm2keygen/users.yaml"
	users, err := loadConfig(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	generateKeys(dirKeys, users)
}

func deleteDir(dirKeys string) error {
	if _, err := os.Stat(dirKeys); os.IsExist(err) {
		err := os.RemoveAll(dirKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadConfig(filePath string) ([]string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config file")
	}

	users := []string{}
	if err = yaml.Unmarshal(fileBytes, &users); err != nil {
		return nil, errors.Wrap(err, "cannot load config file")
	}

	return users, nil
}
