package main

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func loadConfig(filePath string) (config *Config, err error) {
	fileDesc, err := os.Open(filePath)
	if err != nil {
		err = errors.Wrapf(err, "cannot open file '%v'", filePath)
		return
	}
	defer fileDesc.Close()

	configBytes, err := ioutil.ReadAll(fileDesc)
	if err != nil {
		err = errors.Wrapf(err, "cannot read file '%v'", filePath)
		return
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		err = errors.Wrapf(err, "cannot parse config from file '%v'", filePath)
		return
	}

	return
}

type Config struct {
	Tasks []*TaskEntry `yaml:"tasks"`
}

type TaskEntry struct {
	Before string `yaml:"before"`
	After  string `yaml:"after"`
	Diff   string `yaml:"diff"`
}
