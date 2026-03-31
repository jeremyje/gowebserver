// Copyright 2026 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gowebserver

import (
	"context"
	"io/fs"
)

type searchParams struct {
	servePath  string
	indexPath  string
	ollamaSpec string
}

type searchFS interface {
	Start(context.Context) error
	Stop() error
}

func newSearchFS(fsys fs.FS, params *searchParams) searchFS {
	return newNullSearchFS(fsys)
}

type nullSearchFS struct {
	baseFS fs.FS
}

func (nfs *nullSearchFS) FS() fs.FS {
	return nfs.baseFS
}

func newNullSearchFS(fsys fs.FS) searchFS {
	return &nullSearchFS{
		baseFS: fsys,
	}
}

type indexingThread struct {
	sFS        searchFS
	ctx        context.Context
	cancelFunc context.CancelCauseFunc
}

func (th *indexingThread) run(ctx context.Context, cancelFunc context.CancelCauseFunc) error {
	err := fs.WalkDir(th.sFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	})
	return err

}

func (th *indexingThread) Start(ctx context.Context) error {
	if th.ctx != nil {
		return nil
	}

	ctx, cancel := context.WithCancelCause(ctx)

	th.ctx = ctx
	th.cancelFunc = cancel
	go func() {
		th.run(ctx, cancel)
	}()
	return nil
}

func (th *indexingThread) Stop() error {
	if th.ctx != nil {
		th.cancelFunc(nil)
		th.ctx = nil
	}
	return nil
}
