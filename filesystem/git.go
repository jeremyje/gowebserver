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
	"fmt"
	git "gopkg.in/src-d/go-git.v4"
	"os"
	"path/filepath"
	"strings"
)

func newGitFs(filePath string) createFsResult {
	staged := createFsResult{
		localFilePath: filePath,
	}
	if !isSupportedGit(filePath) {
		return staged.withError(fmt.Errorf("%s is not a valid git repository", filePath))
	}

	tmpDir, err := createTempDirectory()
	if err != nil {
		return staged.withError(fmt.Errorf("cannot create temp directory, %s", err))
	}
	staged.tmpDir = tmpDir
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:          filePath,
		Progress:     os.Stdout,
		Depth:        1,
		SingleBranch: true,
	})
	if err != nil {
		return staged.withError(fmt.Errorf("could not clone %s, %s", filePath, err))
	}
	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	return staged.withHandler(newNative(tmpDir))
}

func isSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}
