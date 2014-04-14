package main

import (
	"gopkg.in/yaml.v1"
	"io/ioutil"
)

type Configuration struct {
	Files        []string
	Detect       [][]string
	Excerpt_size int
	Mail         Mail
}

func (config *Configuration) parseConfig() {
	file, err := ioutil.ReadFile("config.yaml")
	check(err)

	err = yaml.Unmarshal(file, &config)
	check(err)
}
