package filesystem

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func newZipFs(filePath string) (http.FileSystem, string, string, error) {
	localFilePath, tmpDir, err := stageRemoteFile(filePath)
	if err != nil {
		return nil, localFilePath, tmpDir, err
	}
	// Extract archive
	r, err := zip.OpenReader(localFilePath)
	if err != nil {
		return nil, localFilePath, tmpDir, err
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		filePath := filepath.Join(tmpDir, f.Name)
		if f.FileInfo().IsDir() {
			err = createDirectory(filePath)
			if err != nil {
				log.Fatalf("Cannot create directory: %s, Error= %v", filePath, err)
				return nil, localFilePath, tmpDir, err
			}
		} else {
			dirPath := filepath.Dir(filePath)
			err = createDirectory(dirPath)
			if err != nil {
				log.Fatalf("Cannot create directory: %s, Error= %v", dirPath, err)
				return nil, localFilePath, tmpDir, err
			}

			err := writeFileFromZipEntry(f, filePath)
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
		}
	}
	handler, err := newNative(tmpDir)
	return handler, localFilePath, tmpDir, err
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
