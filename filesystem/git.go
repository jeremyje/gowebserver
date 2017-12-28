package filesystem

import (
	"fmt"
	git "gopkg.in/src-d/go-git.v4"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func newGitFs(filePath string) (http.FileSystem, string, string, error) {
	if !isSupportedGit(filePath) {
		return nil, "", "", fmt.Errorf("%s is not a valid git repository", filePath)
	}

	tmpDir, err := createTempDirectory()
	if err != nil {
		return nil, "", "", err
	}
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:          filePath,
		Progress:     os.Stdout,
		Depth:        1,
		SingleBranch: true,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("could not clone %s, %s", filePath, err)
	}
	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	handler, err := newNative(tmpDir)
	return handler, filePath, tmpDir, err
}

func isSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}
