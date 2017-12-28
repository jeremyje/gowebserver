package filesystem

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func newTarFs(filePath string) (http.FileSystem, string, string, error) {
	if !isSupportedTar(filePath) {
		return nil, "", "", fmt.Errorf("%s is not a valid tarball", filePath)
	}
	localFilePath, tmpDir, err := stageRemoteFile(filePath)
	if err != nil {
		return nil, localFilePath, tmpDir, err
	}

	var r io.Reader
	f, err := os.Open(localFilePath)
	if err != nil {
		return nil, localFilePath, tmpDir, err
	}
	defer f.Close()
	r = f

	if isTarGzip(filePath) {
		gzf, err := gzip.NewReader(f)
		if err != nil {
			return nil, localFilePath, tmpDir, err
		}
		r = gzf
	} else if isTarBzip2(filePath) {
		bzf := bzip2.NewReader(f)
		r = bzf
	}
	tr := tar.NewReader(r)

	err = processTarEntries(tr, tmpDir)
	if err != nil {
		return nil, localFilePath, tmpDir, err
	}
	handler, err := newNative(tmpDir)
	return handler, localFilePath, tmpDir, err
}

func processTarEntries(tr *tar.Reader, tmpDir string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Cannot get next tar entry, %s", err)
		}

		localPath := filepath.Join(tmpDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = createDirectory(localPath)
		case tar.TypeReg:
			err = writeFileFromTarEntry(localPath, tmpDir, tr)
		default:
			log.Printf("WARNING: Tar entry type not supported. Name= %s, Header= %v", header.Name, header)
		}
		if err != nil {
			return fmt.Errorf("Error processing tar entry %v, %s", header.Typeflag, err)
		}
	}
	return nil
}

func writeFileFromTarEntry(localPath string, tmpDir string, tr *tar.Reader) error {
	err := createDirectory(filepath.Dir(localPath))
	if err != nil {
		return err
	}
	return copyFile(tr, localPath)
}

func isSupportedTar(filePath string) bool {
	return isRegularTar(filePath) || isTarGzip(filePath) || isTarBzip2(filePath)
}

func isRegularTar(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar")
}

func isTarGzip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar.gz")
}

func isTarBzip2(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar.bz2")
}
