package testing

import (
	"encoding/base64"
	"io/ioutil"
	"os"
)

const FILE_MODE = os.FileMode(0644)

func CreateTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}

func WriteTempFile(content string) (*os.File, error) {
	fp, err := CreateTempFile()
	if err != nil {
		return fp, err
	}
	err = ioutil.WriteFile(fp.Name(), []byte(content), FILE_MODE)
	return fp, err
}

func GetZipFilePath() (string, error) {
	archivePath, err := createTempArchive(".zip")
	if err != nil {
		return "", err
	}
	data, err := GetZipContents()
	if err != nil {
		return "", err
	}
	return writeData(archivePath, data)
}

func GetZipContents() ([]byte, error) {
	return base64.StdEncoding.DecodeString(ZIP_ASSETS)
}

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
	return base64.StdEncoding.DecodeString(TAR_ASSETS)
}

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
	return base64.StdEncoding.DecodeString(TAR_GZ_ASSETS)
}

// GetTarBzip2FilePath gets .tar.gz test asset file.
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
	return base64.StdEncoding.DecodeString(TAR_BZIP2_ASSETS)
}

func createTempArchive(suffix string) (string, error) {
	tf, err := CreateTempFile()
	if err != nil {
		return "", err
	}
	return tf.Name() + suffix, nil
}

func writeData(path string, data []byte) (string, error) {
	err := ioutil.WriteFile(path, data, FILE_MODE)
	return path, err
}
