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

package gowebserver

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const emptyConfigYaml = `verbose: false
serve: []
uploadPath: ""
uploadHTTPPath: ""
http:
  port: 0
https:
  port: 0
  certificate:
    rootPrivateKey: ""
    rootPath: ""
    privateKey: ""
    path: ""
    hosts: ""
    duration: 0
metrics:
  enabled: false
  path: ""
`

const populatedConfigYaml = `verbose: true
serve:
- source: /home/folder
  httpPath: /serving
uploadPath: /home/upload
uploadHTTPPath: /postage
http:
  port: 1000
https:
  port: 2000
  certificate:
    rootPrivateKey: root-private-key.pem
    rootPath: root-public-certificate.pem
    privateKey: private-key.pem
    path: public-certificate.pem
    hosts: gowebserver.com
    duration: 9000
metrics:
  enabled: true
  path: /prometheus
`

func TestEmptyConfig(t *testing.T) {
	conf := &Config{}

	if diff := cmp.Diff(emptyConfigYaml, conf.String()); diff != "" {
		t.Errorf("config.String() mismatch (-want +got):\n%s", diff)
	}
}

func TestPopulatedConfig(t *testing.T) {
	conf := &Config{
		Verbose: true,
		Serve: []Serve{{
			Source:   "/home/folder",
			HTTPPath: "/serving",
		}},
		UploadPath:     "/home/upload",
		UploadHTTPPath: "/postage",
		HTTP: HTTP{
			Port: 1000,
		},
		HTTPS: HTTPS{
			Port: 2000,
			Certificate: Certificate{
				RootPrivateKeyFilePath:   "root-private-key.pem",
				RootCertificateFilePath:  "root-public-certificate.pem",
				PrivateKeyFilePath:       "private-key.pem",
				CertificateFilePath:      "public-certificate.pem",
				CertificateHosts:         "gowebserver.com",
				CertificateValidDuration: 9000,
				ForceOverwrite:           true,
			},
		},
		Metrics: Metrics{
			Enabled: true,
			Path:    "/prometheus",
		},
	}

	if diff := cmp.Diff(populatedConfigYaml, conf.String()); diff != "" {
		t.Errorf("config.String() mismatch (-want +got):\n%s", diff)
	}
}

const noDefaultsConfigFile = `verbose: true
serve:
- source: "/home/example"
  httpPath: "/serving"
configurationfile: "/something.yaml"
http:
  port: 1
https:
  port: 2
  certificate:
    privateKey: private.pem
    path: public.pem
    rootPrivateKey: root-private.pem
    rootPath: root-public.pem
    hosts: "hosts"
    duration: 1
    forceoverwrite: false
metrics:
  enabled: false
  path: /metrics
uploadPath: "dropsite"
uploadHTTPPath: "/upload.jspx"
`

func TestNoDefaultConfig(t *testing.T) {
	fp, err := writeTempFile(noDefaultsConfigFile)
	defer os.Remove(fp.Name())
	if err != nil {
		t.Fatal(err)
	}

	got := &Config{}
	err = loadWithConfigFile(fp.Name(), got)
	if err != nil {
		t.Fatal(err)
	}

	want := &Config{
		Verbose: true,
		Serve: []Serve{
			{
				Source:   "/home/example",
				HTTPPath: "/serving",
			},
		},
		ConfigurationFile: "",
		HTTP: HTTP{
			Port: 1,
		},
		HTTPS: HTTPS{
			Port: 2,
			Certificate: Certificate{
				RootPrivateKeyFilePath:   "root-private.pem",
				RootCertificateFilePath:  "root-public.pem",
				PrivateKeyFilePath:       "private.pem",
				CertificateFilePath:      "public.pem",
				CertificateHosts:         "hosts",
				CertificateValidDuration: 1,
				ForceOverwrite:           false,
			},
		},
		Metrics: Metrics{
			Enabled: false,
			Path:    "/metrics",
		},
		UploadPath:     "dropsite",
		UploadHTTPPath: "/upload.jspx",
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("config mismatch (-want +got):\n%s", diff)
	}
}

func TestPopulatedYamlConfig(t *testing.T) {
	fp, err := writeTempFile(populatedConfigYaml)
	defer os.Remove(fp.Name())
	if err != nil {
		t.Fatal(err)
	}

	got := &Config{}
	err = loadWithConfigFile(fp.Name(), got)
	if err != nil {
		t.Fatal(err)
	}

	want := &Config{
		Verbose: true,
		Serve: []Serve{
			{
				Source:   "/home/folder",
				HTTPPath: "/serving",
			},
		},
		ConfigurationFile: "",
		HTTP: HTTP{
			Port: 1000,
		},
		HTTPS: HTTPS{
			Port: 2000,
			Certificate: Certificate{
				RootPrivateKeyFilePath:   "root-private-key.pem",
				RootCertificateFilePath:  "root-public-certificate.pem",
				PrivateKeyFilePath:       "private-key.pem",
				CertificateFilePath:      "public-certificate.pem",
				CertificateHosts:         "gowebserver.com",
				CertificateValidDuration: 9000,
				ForceOverwrite:           false,
			},
		},
		Metrics: Metrics{
			Enabled: true,
			Path:    "/prometheus",
		},
		UploadPath:     "/home/upload",
		UploadHTTPPath: "/postage",
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("config mismatch (-want +got):\n%s", diff)
	}
}

func createTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}

func writeTempFile(content string) (*os.File, error) {
	fp, err := createTempFile()
	if err != nil {
		return fp, err
	}
	err = ioutil.WriteFile(fp.Name(), []byte(content), os.FileMode(0644))
	return fp, err
}

func TestDefaultConfiguration(t *testing.T) {
	got, err := loadFromFlags()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("got is nil")
	}

	want := &Config{
		Verbose: false,
		Serve: []Serve{
			{
				Source:   "",
				HTTPPath: "/",
			},
		},
		ConfigurationFile: "",
		HTTP: HTTP{
			Port: *httpPortFlag,
		},
		HTTPS: HTTPS{
			Port: *httpsPortFlag,
			Certificate: Certificate{
				RootPrivateKeyFilePath:   "",
				RootCertificateFilePath:  "",
				PrivateKeyFilePath:       "web.key",
				CertificateFilePath:      "web.cert",
				CertificateHosts:         "",
				CertificateValidDuration: 5475,
				ForceOverwrite:           false,
			},
		},
		Metrics: Metrics{
			Enabled: true,
			Path:    "/metrics",
		},
		UploadPath:     "uploaded-files",
		UploadHTTPPath: "/upload.asp",
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("config mismatch (-want +got):\n%s", diff)
	}
}
