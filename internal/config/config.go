package config

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	API_KEY    string `yaml:"api_key"`
	ApiKeyPath string `yaml:"api_key_path"`
	CachePath  string `yaml:"cache_dir"`
	MaxResults int64  `yaml:"max_results"`
	Region     string `yaml:"region"`
	ThumbOff   bool   `yaml:"thumbnails_disable"`
	ThumbSize  string `yaml:"thumbnails_size"`
}

var confInstance AppConfig
var once sync.Once
var unmarshalError error

func GetAppConfig(path string) (AppConfig, error) {
	once.Do(func() {
		confInstance = AppConfig{}
		unmarshalError = yaml.Unmarshal(readFile(path), &confInstance)
		if unmarshalError != nil {
			log.Printf("config unmarshal error: %v\n", unmarshalError)
		}
	})

	return confInstance, unmarshalError
}

func readFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("config load error %v\n", err)
	}
	return data
}
