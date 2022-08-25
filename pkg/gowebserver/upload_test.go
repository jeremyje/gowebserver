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
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
	"go.uber.org/zap"
)

func TestUploadHTML(t *testing.T) {
	if len(uploadHTML) < 50 {
		t.Errorf("data/upload.html was not stored")
	}
}

func TestUpload(t *testing.T) {
	zipPath := gowsTesting.MustZipFilePath(t)
	tmpDir, close, err := createTempDirectory()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	cfg := &Config{
		Serve: []Serve{
			{
				Source:   tmpDir,
				Endpoint: "/",
			},
		},
		Upload: Serve{
			Source:   tmpDir,
			Endpoint: "/upload",
		},
	}

	baseURL, close := serveAsync(t, cfg)
	defer close()
	fp, err := os.Open(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	req, err := newUploadFormRequest(ctx, baseURL+"/upload", "test.zip", fp, map[string]string{})
	if err != nil {
		t.Error(err)
	}
	hc := &http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("http status code is '%d'", resp.StatusCode)
	}

	got := sha256File(t, filepath.Join(tmpDir, "test.zip"))
	want := sha256File(t, zipPath)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("sha256 upload mismatch (-want +got):\n%s", diff)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/upload", nil)
	if err != nil {
		t.Error(err)
	}
	resp, err = hc.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("http status code is '%d'", resp.StatusCode)
	}
}

func sha256File(tb testing.TB, localPath string) string {
	f, err := os.Open(localPath)
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		tb.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// https://matt.aimonetti.net/posts/2013-07-golang-multipart-file-upload-example/
func newUploadFormRequest(ctx context.Context, requestURL string, fileName string, reader io.Reader, formEntries map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(uploadFileFormName, fileName)
	if err != nil {
		return nil, err
	}

	bytesRead, err := io.Copy(part, reader)
	if err != nil {
		return nil, err
	}
	zap.S().Infof("bytes read: %d", bytesRead)

	for k, v := range formEntries {
		err = writer.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
