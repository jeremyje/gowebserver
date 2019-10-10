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

package cert

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteDefaultCertificate(t *testing.T) {
	assert := assert.New(t)

	certFp, err := createTempFile()
	defer os.Remove(certFp.Name())
	assert.Nil(err)
	certPath := certFp.Name()
	keyFp, err := createTempFile()
	defer os.Remove(keyFp.Name())
	assert.Nil(err)
	keyPath := keyFp.Name()

	certBuilder := NewCertificateBuilder()
	err = certBuilder.WriteCertificate(certPath)
	assert.Nil(err)
	err = certBuilder.WritePrivateKey(keyPath)
	assert.Nil(err)
	cert, err := ReadCertificateFromFile(certPath)
	assert.Nil(err)
	assert.Equal("US", cert.Subject.Country[0])
	assert.Equal("Seattle", cert.Subject.Locality[0])
	assert.Equal("Washington", cert.Subject.Province[0])
}

func createTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}
