package filesystem

import (
	"archive/zip"
	"github.com/jeremyje/gowebserver/termhook"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func newZipFs(filePath string) (http.FileSystem, string, string, error) {
	localFilePath, err := downloadFile(filePath)
	if err != nil {
		return nil, "", "", err
	}
	tmpDir, err := createTempDirectory()
	if err != nil {
		return nil, "", "", err
	}

	termhook.Add(func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Fatalf("Cannot delete directory: %s, Error= %v", tmpDir, err)
		}
	})

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
			err = os.MkdirAll(dirPath(filePath), os.FileMode(0777))
			if err != nil {
				log.Fatalf("Cannot create directory: %s, Error= %v", dirPath, err)
				return nil, localFilePath, tmpDir, err
			}
		} else {
			dirPath := dirPath(filepath.Dir(filePath))
			err = os.MkdirAll(dirPath, os.FileMode(0777))
			if err != nil {
				log.Fatalf("Cannot create directory: %s, Error= %v", dirPath, err)
				return nil, localFilePath, tmpDir, err
			}
			
			zf, err := f.Open()
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
			defer zf.Close()
			fsf, err := os.Create(filePath)
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
			defer fsf.Close()
	
			_, err = io.Copy(fsf, zf)
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
		}
	}
	handler, err := newNative(tmpDir)
	return handler, localFilePath, tmpDir, err
}
