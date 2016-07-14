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
		return nil, "", "", fmt.Errorf("%s is not a valid tarball.", filePath)
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

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, localFilePath, tmpDir, err
		}
		localPath := filepath.Join(tmpDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = createDirectory(localPath)
		case tar.TypeReg:
			fsf, err := os.Create(localPath)
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
			defer fsf.Close()

			_, err = io.Copy(fsf, tr)
			if err != nil {
				return nil, localFilePath, tmpDir, err
			}
		default:
			log.Printf("WARNING: Tar entry type not supported. Name= %s, Header= %v", header.Name, header)
		}
	}
	handler, err := newNative(tmpDir)
	return handler, localFilePath, tmpDir, err
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
