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
