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

package filesystem

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"
)

func newZipFs(filePath string) createFsResult {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}
	// Extract archive
	r, err := zip.OpenReader(staged.localFilePath)
	if err != nil {
		return staged.withError(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		filePath := filepath.Join(staged.tmpDir, f.Name)
		if f.FileInfo().IsDir() {
			err = createDirectory(filePath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot create directory: %s, %s", filePath, err))
			}
		} else {
			dirPath := filepath.Dir(filePath)
			err = createDirectory(dirPath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot create directory: %s, %s", dirPath, err))
			}

			err := writeFileFromZipEntry(f, filePath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot write zip file entry: %s, %s", f.Name, err))
			}
		}
	}
	return staged.withHandler(newNative(staged.tmpDir))
}

func writeFileFromZipEntry(f *zip.File, filePath string) error {
	zf, err := f.Open()
	if err != nil {
		return fmt.Errorf("Cannot open input file: %s", err)
	}
	defer zf.Close()
	return copyFile(zf, filePath)
}

func isSupportedZip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".zip")
}
