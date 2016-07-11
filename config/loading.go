package config

import (
    "fmt"
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func Load() *Config {
    flag.Parse()
    conf := Get()
    if *configFileFlag != "" {
        contents, err := ioutil.ReadFile(*configFileFlag)
        if err != nil {
            fmt.Printf("Error Loading File: %s", err)
            return conf
        }
        err = yaml.Unmarshal(contents, &conf)
        if err != nil {
            fmt.Printf("Error Loading File: %s", err)
            return conf
        }
    }
    return conf
}
