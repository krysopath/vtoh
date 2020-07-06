package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type FileBackend struct {
	FilePath string
}

func (backend FileBackend) Load() ([]byte, error) {
	content := make([]byte, 0)
	if _, err := os.Stat(backend.FilePath); err == nil {
		content, err = ioutil.ReadFile(backend.FilePath)
		if err != nil {
			panic(err)
		}
	}
	return content, nil
}

func (backend FileBackend) Save(data interface{}) (bool, error) {
	dataBytes, yamlErr := yaml.Marshal(data)
	if yamlErr != nil {
		panic(yamlErr)
	}
	writeErr := ioutil.WriteFile(
		backend.FilePath,
		dataBytes,
		0400)

	if writeErr != nil {
		panic(writeErr)
	}
	return true, nil
}
