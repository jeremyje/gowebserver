package config

import (
	"flag"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
)

func Load() *Config {
	flag.Parse()
	conf := loadFromFlags()
	if *configFileFlag != "" {
		err := loadWithConfigFile(*configFileFlag, conf)
		if err != nil {
			fmt.Printf("Error Loading File: %s", err)
		}
	}
	return conf
}

func loadWithConfigFile(filePath string, conf *Config) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contents, &conf)
	// Config file should always be the flag value.
	conf.ConfigurationFile = *configFileFlag
	if err != nil {
		return err
	}
	return nil
}
