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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func newGitFS(filePath string) (*localFS, error) {
	if !isSupportedGit(filePath) {
		return nil, fmt.Errorf("%s is not a valid git repository", filePath)
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("cannot create temp directory, %w", err)
	}

	for _, opts := range cloneOptions(filePath) {
		if _, err = git.PlainClone(tmpDir, false, opts); err == nil {
			break
		}
	}
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("could not clone %s, %w", filePath, err)
	}

	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	return newLocalFS(tmpDir, func() error {
		cleanup()
		return nil
	})
}

func isSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}

func cloneOptions(filePath string) []*git.CloneOptions {
	return []*git.CloneOptions{
		{
			URL:          filePath,
			Progress:     os.Stdout,
			Depth:        1,
			SingleBranch: true,
		},
		{
			URL:           filePath,
			Progress:      os.Stdout,
			Depth:         1,
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName("main"),
		},
	}
}
