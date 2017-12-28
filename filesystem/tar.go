package filesystem

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func newTarFs(filePath string) createFsResult {
	if !isSupportedTar(filePath) {
		return createFsResult{
			err: fmt.Errorf("%s is not a valid tarball", filePath),
		}
	}
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}

	var r io.Reader
	f, err := os.Open(staged.localFilePath)
	if err != nil {
		return staged.withError(err)
	}
	defer f.Close()
	r = f

	if isTarGzip(filePath) {
		gzf, err := gzip.NewReader(f)
		if err != nil {
			return staged.withError(err)
		}
		r = gzf
	} else if isTarBzip2(filePath) {
		bzf := bzip2.NewReader(f)
		r = bzf
	}
	tr := tar.NewReader(r)

	err = processTarEntries(tr, staged.tmpDir)
	if err != nil {
		return staged.withError(err)
	}
	return staged.withHandler(newNative(staged.tmpDir))
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
