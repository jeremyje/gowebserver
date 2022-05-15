package testing

import (
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	defaultFileData = []byte("ok")
)

func TestCreateTempFile(t *testing.T) {
	testCases := []struct {
		name string
		f    func(testing.TB) (string, error)
	}{
		{
			name: "MustCreateTempFile",
			f: func(tb testing.TB) (string, error) {
				fp := MustCreateTempFile(tb)
				n, err := fp.Write(defaultFileData)
				if err != nil {
					return "", err
				}
				if n != len(defaultFileData) {
					t.Errorf("bytes written, got %d, want %d", n, len(defaultFileData))
				}

				if err := fp.Close(); err != nil {
					return "", err
				}
				return fp.Name(), nil
			},
		},
		{
			name: "MustWriteTempFile",
			f: func(tb testing.TB) (string, error) {
				fp := MustWriteTempFile(tb, string(defaultFileData))
				return fp.Name(), nil
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			name, err := tc.f(t)
			if err != nil {
				t.Fatal(err)
			}

			data, err := ioutil.ReadFile(name)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(data, defaultFileData); diff != "" {
				t.Errorf("ExpandHostnames() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMustFilePath(t *testing.T) {
	testCases := []struct {
		name string
		f    func(testing.TB) string
	}{
		{name: "MustZipFilePath", f: MustZipFilePath},
		{name: "MustSevenZipFilePath", f: MustSevenZipFilePath},
		{name: "MustTarFilePath", f: MustTarFilePath},
		{name: "MustTarGzFilePath", f: MustTarGzFilePath},
		{name: "MustTarBzip2FilePath", f: MustTarBzip2FilePath},
		{name: "MustTarXzFilePath", f: MustTarXzFilePath},
		{name: "MustTarLz4FilePath", f: MustTarLz4FilePath},
		{name: "MustZipFilePath", f: MustZipFilePath},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			name := tc.f(t)
			if name == "" {
				t.Error("file name is empty")
			}
			data, err := ioutil.ReadFile(name)
			if err != nil {
				t.Error(err)
			}
			if len(data) == 0 {
				t.Error("file contents are empty")
			}
		})
	}
}
