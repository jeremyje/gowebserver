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

package testing

import (
	"encoding/base64"
	"io/ioutil"
	"os"
)

const testFileMode = os.FileMode(0644)

// CreateTempFile creates a temp file for testing.
func CreateTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}

// WriteTempFile writes the contents to a file for testing.
func WriteTempFile(content string) (*os.File, error) {
	fp, err := CreateTempFile()
	if err != nil {
		return fp, err
	}
	err = ioutil.WriteFile(fp.Name(), []byte(content), testFileMode)
	return fp, err
}

// GetZipFilePath gets the .zip test asset file.
func GetZipFilePath() (string, error) {
	archivePath, err := createTempArchive(".zip")
	if err != nil {
		return "", err
	}
	data, err := getZipContents()
	if err != nil {
		return "", err
	}
	return writeData(archivePath, data)
}

func getZipContents() ([]byte, error) {
	return base64.StdEncoding.DecodeString(zipAssets)
}

// GetTarFilePath gets the .tar test asset file.
func GetTarFilePath() (string, error) {
	archivePath, err := createTempArchive(".tar")
	if err != nil {
		return "", err
	}
	data, err := getTarContents()
	if err != nil {
		return "", err
	}
	return writeData(archivePath, data)
}

func getTarContents() ([]byte, error) {
	return base64.StdEncoding.DecodeString(tarAssets)
}

// GetTarGzFilePath gets the .tar.gz test asset file.
func GetTarGzFilePath() (string, error) {
	archivePath, err := createTempArchive(".tar.gz")
	if err != nil {
		return "", err
	}
	data, err := getTarGzContents()
	if err != nil {
		return "", err
	}
	return writeData(archivePath, data)
}

func getTarGzContents() ([]byte, error) {
	return base64.StdEncoding.DecodeString(tarGzAssets)
}

// GetTarBzip2FilePath gets .tar.bz2 test asset file.
func GetTarBzip2FilePath() (string, error) {
	archivePath, err := createTempArchive(".tar.bz2")
	if err != nil {
		return "", err
	}
	data, err := getTarBzip2Contents()
	if err != nil {
		return "", err
	}
	return writeData(archivePath, data)
}

func getTarBzip2Contents() ([]byte, error) {
	return base64.StdEncoding.DecodeString(tarBzip2Assets)
}

func createTempArchive(suffix string) (string, error) {
	tf, err := CreateTempFile()
	if err != nil {
		return "", err
	}
	return tf.Name() + suffix, nil
}

func writeData(path string, data []byte) (string, error) {
	err := ioutil.WriteFile(path, data, testFileMode)
	return path, err
}
