package testing

import (
	"encoding/base64"
	"io/ioutil"
	"os"
)

func CreateTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}

func WriteTempFile(content string) (*os.File, error) {
	fp, err := createTempFile()
	if err != nil {
		return fp, err
	}
	err = ioutil.WriteFile(fp.Name(), []byte(content), os.FileMode(0644))
	return fp, err
}

func GetZipFilePath() (string, error) {
	tf, err := CreateTempFile()
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(GetZipContents(), tf.Name())
	return tf.Name(), err
}

func GetZipContents() []byte {
	return base64.StdEncoding.DecodeString(ZIP_FILE)
}
