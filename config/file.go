// Copyright 2019 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"flag"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Load loads the configuration for the server.
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
