// Copyright 2022 Jeremy Edwards
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

// Package main is the entry point for certtool.
package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jeremyje/gowebserver/pkg/certtool"
	"go.uber.org/zap"
)

var (
	publicCertificate = flag.String("public-certificate", "app.cert", "X.509 public certificate file to generate.")
	privateKey        = flag.String("private-key", "app.key", "Private key file to generate.")

	ca = flag.Bool("ca", false, "Generates a root certificate. Use this to establish a chain of trust with derived certificates.")

	country            = flag.String("country", "US", "CountryName field of the certificate attribute.")
	organization       = flag.String("organization", "gowebserver", "CountryName field of the certificate attribute.")
	organizationalUnit = flag.String("organizational-unit", "gows", "CountryName field of the certificate attribute.")
	locality           = flag.String("locality", "Seattle", "CountryName field of the certificate attribute.")
	province           = flag.String("province", "WA", "CountryName field of the certificate attribute.")

	hostnames = flag.String("hostnames", "", "Comma separated list of hostnames.")
	keyType   = flag.String("key-type", "RSA-2048", "Type of key to generate. (default: RSA-2048)")

	parentPublicCertificate = flag.String("parent-public-certificate", "", "(optional) Parent public certificate. If set, the output certificate will trust the parent.")
	parentPrivateKey        = flag.String("parent-private-key", "", "(optional) Parent private key. Required if -parent-public-certificate is set, private key for the parent public certificate.")
)

func main() {
	certtoolMain()
}

func certtoolMain() {
	flag.Parse()

	if args, err := argsFromFlags(); err == nil {
		if _, err := certtool.GenerateAndWriteKeyPair(args, *publicCertificate, *privateKey); err != nil {
			zap.S().Error(err)
		}
	} else {
		zap.S().Error(err)
	}
}

func argsFromFlags() (*certtool.Args, error) {
	var parent *certtool.KeyPair

	algorithm, keyLength, err := StringToKeyType(*keyType)
	if err != nil {
		return nil, err
	}

	if *parentPublicCertificate != "" {
		parent, err = certtool.ReadKeyPairFromFile(*parentPublicCertificate, *parentPrivateKey)
		if err != nil {
			return nil, err
		}
	}

	return &certtool.Args{
		CA:                 *ca,
		Country:            *country,
		Organization:       *organization,
		OrganizationalUnit: *organizationalUnit,
		Locality:           *locality,
		Province:           *province,

		Hostnames: ExpandHostnames(*hostnames),
		KeyType: &certtool.KeyType{
			Algorithm: algorithm,
			KeyLength: keyLength,
		},
		ParentKeyPair: parent,
	}, nil
}

func StringToKeyType(keyType string) (string, int, error) {
	if keyType == "" {
		return "RSA", 2048, nil
	}
	switch parseKeyTypeKeyName(keyType) {
	case "RSA":
		return parseKeyTypeName(keyType, 2048, []int{2048, 4096})
	case "ECDSA":
		return parseKeyTypeName(keyType, 521, []int{224, 256, 384, 521})
	}
	return "", 0, fmt.Errorf("'%s' is not a valid key type", keyType)
}

func parseKeyTypeKeyName(keyTypeName string) string {
	parts := strings.Split(strings.ToUpper(keyTypeName), "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "RSA"
}

func parseKeyTypeName(keyTypeName string, defaultLength int, validValues []int) (string, int, error) {
	parts := strings.Split(strings.ToUpper(keyTypeName), "-")
	if len(parts) > 2 {
		return "", 0, fmt.Errorf("key type '%s' is not valid", keyTypeName)
	}
	if len(parts) == 0 {
		return "", 0, fmt.Errorf("key type does not have a name")
	}

	algorithm := parts[0]
	keyLength := ""
	if len(parts) == 2 {
		keyLength = parts[1]
	}

	if keyLength == "" {
		return algorithm, defaultLength, nil
	}
	length, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("key type '%s' does not have a valid %s key length", keyTypeName, algorithm)
	}
	for _, validLength := range validValues {
		if validLength == length {
			return algorithm, length, nil
		}
	}

	return "", 0, fmt.Errorf("key type '%s' does not have a valid %s key length", keyTypeName, algorithm)
}

func ExpandHostnames(hostnameCsv string) []string {
	return expandHostnames(strings.Split(hostnameCsv, ","))
}

func expandHostnames(hostnames []string) []string {
	unique := map[string]interface{}{}

	for _, hn := range hostnames {
		if hn != "" {
			unique[hn] = nil
		}
	}

	all := []string{}

	for hn := range unique {
		all = append(all, hn)
	}
	sort.Strings(all)
	return all
}
