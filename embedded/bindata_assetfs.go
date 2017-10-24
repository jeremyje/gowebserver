// Code generated by go-bindata.
// sources:
// BUILD.bazel
// ajaxupload.html
// basicupload.html
// multipleupload.html
// DO NOT EDIT!

package embedded

import (
	"github.com/elazarl/go-bindata-assetfs"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _buildBazel = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8f\xd1\x6a\xc3\x30\x0c\x45\xdf\xfd\x15\xc6\x4f\x2d\x84\xf8\x3d\x50\xe8\x7f\x8c\x62\xe4\x58\x71\x55\xe4\x2a\xc8\x4e\x47\xf2\xf5\x83\x64\x5b\x61\xec\x51\xe2\x70\xcf\xbd\x2c\x90\x4e\xee\x4a\x12\x22\x6c\xc8\x41\x17\xc6\x1a\xb2\x78\x9f\x65\x48\x38\xf5\x71\x63\xd7\x59\x97\x25\x30\x45\x05\x5d\xdd\xd9\x98\xf7\x75\x32\xd6\x5a\xfb\x84\x82\xf6\xb2\x53\x09\x27\x58\xb8\xfd\xd2\xdd\x0e\x54\x1d\xab\xbd\xd8\x0f\x17\xe9\x99\xa0\x41\x80\x5a\xb1\x4d\xb5\xcf\xe2\x6e\x07\x42\x65\x16\x6d\x33\xb4\xfb\x9e\x44\xed\xbe\xc4\x7e\x94\xe2\x1f\xa8\x58\xd6\x07\xfa\x2c\x9f\x18\x2b\xea\x0b\xd5\x63\x89\x98\x12\xa6\xef\xfc\x17\x55\x8a\xc4\xd4\xd6\xdd\xe2\xfd\xfb\x31\xcc\x4b\x64\x1a\x7f\x34\x09\xe7\xa3\xc9\x75\x94\x12\x0e\x4d\x40\x86\x0d\x94\x43\x96\xf0\xa7\xa0\xf7\xc3\x3f\xa3\x6e\x9d\x39\x9b\xaf\x00\x00\x00\xff\xff\x5a\x9f\x6e\xf0\x3b\x01\x00\x00")

func buildBazelBytes() ([]byte, error) {
	return bindataRead(
		_buildBazel,
		"BUILD.bazel",
	)
}

func buildBazel() (*asset, error) {
	bytes, err := buildBazelBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "BUILD.bazel", size: 315, mode: os.FileMode(420), modTime: time.Unix(1508813733, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ajaxuploadHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x3b\x69\x73\xdb\x38\x96\xdf\xfd\x2b\x5e\xd8\x87\xa4\xb4\x78\x48\xb6\x13\x47\x91\xd4\x9b\x8d\x3b\x35\xb5\xd5\xe9\xe9\xea\x64\xab\x76\x2a\xd3\x3b\x05\x91\x8f\x22\x62\x08\xe0\x00\xd0\xe1\xa4\xfd\xdf\xb7\x00\x90\x14\x2f\xd9\x4e\x36\xa9\x4a\x4c\x02\x78\xf7\x0d\x3a\xf3\x27\xd7\x7f\x7f\xfd\xfe\x1f\xbf\xff\x02\x99\xde\xb0\xe5\xd9\xfc\x89\xef\x43\xa6\x75\xae\x66\x61\x18\x2b\xe5\x6b\x49\xe3\x1b\x15\xc4\x62\x13\xe2\x81\x6c\x72\x86\x2a\xbc\x96\x64\xfd\x8a\x27\xd7\x52\xe4\x6f\x28\xc3\xff\xce\x99\x20\x09\xe5\xeb\x10\x7c\x7f\x79\x36\x37\x98\x80\x11\xbe\x5e\x78\xc8\x3d\x88\x19\x51\x6a\xe1\x71\xe1\x7f\x54\xde\xf2\xec\x6c\x9e\x21\x49\x96\x67\x00\xf3\x0d\x6a\x02\x71\x46\xa4\x42\xbd\xf0\xb6\x3a\xf5\xaf\x3c\xbb\xa1\xa9\x66\xb8\x74\x78\xc1\x90\x50\xf3\xd0\xad\x99\x5d\x46\xf9\x0d\x48\x64\x0b\x2f\x26\x5c\x70\x1a\x13\xe6\x41\x26\x31\x5d\x78\x9f\x3f\x07\x0e\xea\x1d\xca\x1d\xfe\x4e\x74\x76\x77\xe7\x1d\x69\x71\xb2\xc1\x85\xb7\xa3\xb8\xcf\x85\xd4\x1e\xc4\x82\x6b\xe4\x7a\xe1\xed\x69\xa2\xb3\x45\x82\x3b\x1a\xa3\x6f\x5f\xc6\x94\x53\x4d\x09\xf3\x55\x4c\x18\x2e\x26\x1e\x84\x2d\xe2\x4a\xdf\x32\x54\x19\xa2\x2e\xa9\x87\x61\x2a\xb8\x56\xc1\x5a\x88\x35\x43\x92\x53\xa7\xb7\x58\xa9\x9f\x53\xb2\xa1\xec\x76\xf1\x87\x58\x09\x2d\x66\xe7\x51\x34\x3e\x8f\x22\xaa\x09\xa3\xf1\xf8\x22\x8a\x4a\xec\x16\xa7\x79\x02\x6b\x0f\xf8\x7c\x67\x9f\xed\x3f\x2b\x91\xdc\xc2\x67\xfb\x08\x60\x08\xf9\x0e\xe9\x0c\x1c\xd6\x31\x28\xc2\x95\xaf\x50\xd2\xf4\x65\x71\x2c\x16\x4c\xc8\x19\x7c\x17\xa5\xe7\xf1\xc5\xaa\x5c\x5d\x91\xf8\x66\x2d\xc5\x96\x27\x7e\x79\x00\x2f\x31\x49\x27\xe5\x81\x9c\x24\xc6\x9e\x33\xb8\x94\xb8\x81\x49\x30\x35\x3f\xcb\xcd\xf0\x29\x5c\x45\x30\x8d\xe0\x69\x68\x57\x6a\x2c\x06\x46\xa1\x84\x72\x94\x15\xa3\x56\x99\x33\x98\x44\xd1\x0f\x25\x82\x0d\x39\xf8\xc5\xf2\xb3\xab\x28\x3f\x34\x10\x57\x68\x01\x34\x1e\xb4\x4f\x18\x5d\xf3\x19\xc4\xc8\x35\xca\x23\x06\xb9\xa6\x7c\x06\x11\x90\xad\x16\x2f\xef\x61\x23\x9b\x34\x55\xa6\xe8\x27\x9c\xc1\xc5\xf4\x48\xd5\x2e\xef\x91\xae\x33\x3d\x83\xf3\x28\xba\x5f\x75\x8e\xb2\xbf\x12\x5a\x8b\xcd\x0c\x2e\x2a\xf6\x4f\x92\x27\xb3\x4c\xec\x50\x8e\x7b\xb7\x52\x11\x6f\x55\xc5\x61\x49\xf2\xfc\xc5\x2a\x4d\xce\xef\xc3\xcb\xc9\xae\x82\xfa\x42\x96\x0c\x28\xa9\x80\x57\x42\x26\x28\x2b\xe0\x69\x7e\x00\x25\x18\x4d\xe0\xbb\xf8\x2a\x21\x49\xe5\x48\x09\x55\x39\x23\xb7\x33\xa0\x9c\x51\x8e\xfe\x8a\x89\xf8\xa6\xe3\x2e\x17\xf9\x01\xae\x8e\xaa\x3d\x9a\xe9\xf2\x11\x3c\x05\xd4\x38\x2f\xc3\x58\x63\xd2\x34\x5a\x69\x9d\xe7\x5d\xeb\xd4\x55\xd5\x92\xa6\xf4\xed\x78\x2b\x25\x72\xfd\xda\xbc\x3d\xc8\xc4\x8c\x0b\x3d\x84\x3a\x2b\xa3\x13\xf6\x3b\x79\xbc\x69\xd3\x5e\x96\x1a\x4e\xd5\xcf\x4d\x2a\x84\xae\x85\x51\x09\xf8\x62\xba\x8a\x56\xe7\x2d\x6f\xd4\x22\x7f\xd8\xee\x05\xc2\xfc\xa7\xbc\xed\x39\x16\x7c\x52\x06\xf8\xbd\xd0\xa7\x9c\xb9\xda\xfe\x32\x87\x5e\x89\x43\x5f\x78\x76\x13\xce\xb4\x96\x16\x7a\xb2\x57\xd3\x57\x73\xa1\xa8\xa6\x82\xcf\x4c\x9e\x26\x9a\xee\xb0\xe3\xa8\x93\x28\xca\x0f\x30\xed\x55\xd9\x4a\x1c\x82\x8c\x28\x9f\x24\x3b\xc2\x63\x4c\xfc\xad\xab\x43\x25\xa3\x62\xab\x4d\x08\xb8\x60\x49\x88\xca\x30\x69\xdb\xa5\x38\xe2\x8b\x34\x55\xa8\x67\xe0\x4f\x6a\x69\xce\xdf\xe3\xea\x86\x6a\x5f\x4b\xc2\x4b\x46\x9b\x00\x10\x4c\x2e\x15\x20\x51\xe8\x53\xee\x8b\xad\x1e\x77\x84\x76\x47\x0c\x0c\xa9\xb2\xe2\x37\x46\xd8\xd6\x09\x55\x7e\x22\xc9\xda\xd8\xbf\xad\x8b\xa3\xa0\xd3\x9a\xa0\xe5\x66\xbf\x95\x7a\xcc\x98\xa6\x69\x2f\xe9\x7f\xfd\xcb\x10\xe6\x89\x14\xf9\xb8\xb6\x48\x63\xc1\x2b\x4e\xaa\x04\xc5\x05\xc7\xc7\x1b\xb5\x85\xbe\x8b\xce\xe5\xbb\x2f\x45\xd8\x60\xad\xa7\x04\x66\x45\x46\xab\x97\xbf\x94\x32\xd6\x89\xf0\x8a\x8f\x46\xc2\x7d\x54\xca\x2f\x6c\xb6\x2d\xbb\xb3\x92\x35\x9e\x6f\xf5\xb8\x71\x44\x6d\xe3\x18\x95\x3a\x7d\x00\xa5\x34\x1e\x72\xdc\xae\x64\xdb\x51\x45\x57\x94\x51\x7d\x3b\x83\x8c\x26\x09\xf2\x13\x16\xac\xd8\xa8\x5b\xb0\x20\x5c\x5f\x72\xa4\xbe\xc0\xaa\x3d\x22\xf6\xd0\xea\x8a\xd9\xa1\xdd\x16\xf4\x04\x27\xcd\xca\x57\xa5\x1a\xb2\x52\x82\x6d\x75\x95\x6a\x6c\x52\xbd\x3c\x5a\x5c\x3a\x83\x57\x05\x8c\x61\x5a\x7f\x6d\x24\x85\x54\xc8\xcd\xcc\x85\x33\x23\x1a\xff\x31\x04\xff\x32\xfa\x61\xd4\x88\xf3\x7b\xcf\x9c\xd6\x7e\x2b\xdd\x9a\x76\x73\x06\xae\x0f\x3d\x01\x5a\x6a\xed\x73\x8b\x53\xc2\xe9\x86\x14\xb2\xe7\x39\x12\xe9\xa7\x52\x6c\x7c\xca\x15\x4d\x10\x82\x69\x33\xe1\x94\xbc\x7f\x0d\x54\x8d\xa5\xff\x28\xa9\xdf\xe0\x6d\x2a\xc9\x06\x55\x1f\x9a\x4a\x46\x29\x36\xd5\xcb\xe3\x74\x0c\xb6\xdd\x1f\x42\x54\x69\xfb\xae\xf8\xf9\xfc\xf2\x87\xaf\xc4\x35\x09\x26\x1d\x6c\x5a\x7c\x2d\xb2\x16\xaa\xba\x72\xbe\x46\x29\xdf\x42\x19\xdf\x44\x09\xff\x4f\xe1\x9d\xb3\x4a\x54\x9a\x48\xfd\x40\xfb\x78\x0a\xcc\x75\x2f\xe3\x9e\x8d\xac\x51\xf5\x1e\x6a\x6b\x3e\x96\x39\x26\xa5\x0c\xdb\x65\x20\x0a\x26\xc7\x9c\x5f\xd6\x81\xc6\xa2\xc8\x49\x6c\x13\x6a\x95\x1d\x0c\xf5\x94\x89\x7d\x33\xc7\xde\x97\x7e\x3e\xf9\x94\x27\x78\x30\xad\xc7\x03\xfc\xfd\xc4\xc8\x0a\x59\xad\x27\xac\x86\xb3\xab\x63\xf6\xb2\xa3\xd8\x91\x0b\x64\x8c\xe6\x8a\xaa\x72\x7b\x9f\x51\x8d\xbe\xca\x49\x8c\x26\x55\xef\x25\xc9\xab\x46\x7d\x2b\x95\x51\x56\x2e\x68\x7d\x82\xbb\x77\x9e\xe8\x17\xf7\x01\x09\x0a\x1b\x29\x2d\x45\xb3\xc8\x98\x23\xce\xb2\x85\xa8\xfd\x47\x6c\x35\xef\x1e\xfb\x3a\xa3\xd7\xe9\x8d\x7b\xf6\x3b\xc4\x3a\xcd\xe5\xc4\x34\x97\x42\x9b\x59\xe8\xbb\xe8\x38\xf7\x54\xfb\x65\xce\xb0\x48\x7c\x49\xf9\xba\xe8\xe3\xcc\x44\xdc\x3f\x6f\xf5\x99\xfd\x69\x45\x39\x7c\x5a\xda\xc8\xc7\x1d\x72\xad\x8a\xa2\x7b\x6c\xbc\xc3\xa7\x40\x39\xc4\x44\x21\x88\x14\xde\x10\xa5\x5f\x33\x1a\xdf\x00\xa3\x2b\xd8\x2a\xec\xbb\x0f\xb0\x97\x3e\xa7\x9d\xed\x81\xf2\x5e\x87\x5e\x6d\xb5\xee\x6b\xf7\x6a\x8e\xd3\x09\xeb\x16\xcc\x43\xb3\x64\xf3\x0e\xa4\xa7\x3f\x6d\x0e\x9b\x3d\xcc\xd7\x26\x8c\xab\xfc\x00\x93\x67\xdd\x59\xd8\xf4\x6a\xce\x46\xa7\x92\x91\xe3\xba\x31\x68\xd5\xd7\x5b\xe3\x65\x97\xcb\xf6\x6c\x39\x0f\x8b\x5b\x25\xf3\xfc\xc4\xf7\x41\xe2\x46\xec\x10\x74\x46\x15\xd0\x14\x6e\xc5\xd6\x9a\xef\xad\x48\x50\x72\xfa\x49\xda\x5b\x3c\x80\xb9\x8a\x25\xcd\xb5\xbb\x8d\x1a\xa6\x5b\x1e\x9b\x54\x33\xc4\x31\xe8\x31\xf0\xd1\xb1\x05\x24\x12\x24\x2c\x00\x83\x7f\x6f\x51\xde\xbe\xb3\x83\xb0\x90\xaf\x18\x1b\x7a\x99\xde\x30\x6f\xf4\x21\xfa\xb3\x6a\x83\x02\x7b\x1d\xf8\x1b\xd9\x20\x2c\xea\x6f\x81\xc4\x9c\x91\x18\x87\xe1\xf0\x7f\xff\xfa\xa7\x1a\x59\xdb\x0f\xff\xa9\xfe\xfa\x7e\x14\x8e\xc1\xfb\x7e\xf2\x51\x7d\x3f\xf5\x46\x4e\xaa\xd1\x30\x11\xf1\x76\x83\x5c\x8f\x61\x4f\x79\x22\xf6\xe3\xa2\x48\xcd\xc3\x92\xeb\xb3\x79\xe8\x2e\x1a\xcf\xe6\x2b\x91\xdc\x5a\x89\x12\xba\x2b\xaf\x23\xab\xd9\xd5\x03\x29\x18\x2e\xbc\x0d\xa1\xdc\x73\xc2\xce\x4d\x19\x82\x0d\xea\x4c\x24\x0b\x2f\x17\x4a\x7b\x40\xac\xf4\x27\x6e\x18\x01\x79\xac\x6f\x73\x83\x65\xcb\x34\xcd\x89\xd4\xa1\xc1\xe1\x27\x44\x13\x0f\xb8\xd8\x11\x46\x13\xa2\xb1\xa4\xbe\x12\x87\x82\x56\x93\xab\x63\x87\x5d\x6d\x1b\x43\xec\xd6\xcd\x03\xb1\xe0\x1e\x1c\x36\x8c\xab\x85\x97\x69\x9d\xcf\xc2\x70\xbf\xdf\x07\xfb\xf3\x40\xc8\x75\x38\x8d\xa2\x28\x54\xbb\xb5\xe7\x4a\xce\xc2\xbb\x8c\xbc\xa2\xd2\x2c\xbc\x8b\x73\x0f\x76\x14\xf7\xff\x29\x0e\x0b\x2f\x82\x08\x2e\x23\xb8\x38\xf7\x96\xf3\x9c\xe8\x0c\x92\x85\xf7\xf6\xe2\x2a\xb8\x80\xe9\xb3\xe0\x32\xf6\x83\x17\x10\xf9\x93\xe0\x79\xf0\xdc\xfc\x0b\x93\xe0\xf9\x6e\x32\x09\x9e\x65\xfe\xc5\x79\x70\xbe\xf3\xcd\x73\x1c\xf9\xc1\x0b\xdf\x1d\x28\xff\xaa\x36\xcc\x79\x30\x8d\x23\x08\x5e\xb8\xf7\xf2\x6f\x76\xf1\x2c\x78\x1e\x1b\x1a\xe6\xcd\x77\x8b\x06\x68\xe7\x3b\x08\x7f\xd2\x42\xfc\x69\xe3\x4f\x2f\x82\x4b\x78\x16\x4c\xe2\xe0\x3c\x38\x0f\xae\x82\x4b\x98\x04\xd3\xe0\x32\xb8\x00\x43\xc0\x0f\xa6\xe6\xdd\x0f\x2e\xd9\x24\x72\xfc\x59\xcc\x05\x33\x91\x3f\x0d\x2e\x2c\x77\x06\xf1\xd4\x00\x31\xff\x79\x30\x81\x2b\x23\xce\xf4\x32\x38\x7f\x8c\x38\xe6\x9c\x05\xf3\xaf\x82\xf3\xd8\x77\xc8\xea\x38\x55\x21\x0b\x44\x30\x0d\x2e\xd8\x24\x02\xc3\xc9\x27\x2f\x5c\xce\x8d\x5d\x6a\x96\x75\xc3\x94\x73\x1d\x93\x1f\xbd\xe2\x8a\xba\xf2\xb2\x37\x94\xe1\x1b\x21\x37\x26\x44\x8c\xa3\xd1\xa4\x3c\x58\x77\x08\xb7\x62\x9c\xcd\x77\x0e\xc8\xd0\x8f\x49\x5e\x78\x6c\x2c\xb6\x5c\xdf\x99\xf9\x12\x15\x94\xf7\x54\x1e\x94\x27\xdd\xd5\x73\xc1\x90\xcb\xcf\xa9\x90\x05\x99\xe5\xdc\x95\xc1\xe5\xeb\x4c\x08\x85\x40\x2c\x1a\x93\x52\xec\xea\x5c\xe5\x84\x37\x58\xa9\x46\x69\x6f\x09\x42\x82\x79\x05\xaa\x21\x43\x69\xa0\x72\xc2\x97\xc1\x3c\xb4\x54\x6a\x54\x8b\x44\xed\xf4\xa0\xb6\xab\x0d\xd5\x4d\x01\xdd\x01\xaf\xf8\x24\x30\x0f\xdd\x7b\x15\x41\x61\x42\x77\xd5\x4b\x5d\xa7\xae\x7f\x28\xb5\xaa\xc5\x8d\x79\xd9\x11\xb6\xad\xeb\xf8\xbd\x59\x36\xca\x0d\x6d\x8e\xec\x8d\xc9\x6a\x80\x2a\x79\xa0\x7c\xfd\x63\x66\xfb\xa0\x97\x4d\xf2\x6d\xc8\x62\x7e\xf2\x96\xd7\x82\xe3\x13\x98\x93\xfb\x3e\x55\x34\x20\x8b\xe6\xb3\x4c\x50\x4d\x1d\xc0\x46\x48\xfc\x79\x1e\x92\xe5\xfd\xe4\xed\x14\xeb\x2d\x7f\x31\x3f\x9e\x80\xb5\xd7\xb2\x34\xc4\x37\x60\xe6\xbd\xbc\x05\xb2\x26\x94\x3f\x69\xb1\x32\xb7\x09\xb0\xca\xa7\x42\xa3\xac\x78\xcc\x2b\xaf\xba\x36\xee\x41\x78\x02\xd7\x52\xe4\xc6\x5f\x62\xdb\x51\x68\x01\xc5\x75\x8a\x12\x1b\x74\x9e\x1b\x1c\x9d\x2e\xcc\x8f\x34\x4a\xc4\x05\xe9\x76\xdd\x1a\x98\xca\xa6\xb4\xa4\xb1\x1e\xbc\x6c\x55\xb2\x6e\x0d\xb1\x3d\xf3\xb1\xb2\x85\x21\xa4\x48\xf4\x56\x22\x24\xa8\xd1\x42\x99\xd8\xb0\x5e\xfd\xa3\xbd\x2f\x72\x6c\xd6\x2a\x21\x55\xaf\x8a\x1b\xa1\xc2\x4e\x0b\xa8\x08\x8e\x6a\xe3\x8e\x39\x6b\x6c\xb5\x80\x92\x8d\x20\x96\x48\x34\xfe\xc2\xd0\xbc\x0d\x07\x09\xdd\x0d\xaa\x79\x07\x40\xa2\xde\x4a\x0e\xc3\xe1\xc0\x50\x5f\x93\x15\xc3\x81\xe9\xc7\x12\xba\x1b\xc1\x5f\x7f\xc1\x70\x20\xb8\xd9\xb1\x66\x2a\x77\xe0\xc7\x1f\xc1\xae\x8b\xbc\x3a\x3c\xb2\x8b\x26\xa9\x5c\x13\x4d\xec\xb2\x93\xdf\xad\x53\x86\x7f\x20\x49\x50\xd6\x76\xaa\xa9\x6b\x38\x7a\x79\x76\x54\x0e\xc9\x73\x76\x4b\xf9\x1a\x74\x86\x80\x69\x8a\xb1\xb6\xea\xc1\x1d\xca\x5b\xf3\xb4\xa9\x29\xc6\xbc\xaa\xba\xb8\x9d\x6e\x61\x60\xfa\x9c\xa3\xc8\xaf\xa4\x24\xb7\x41\x2e\x85\x16\x26\x98\x83\x54\xc8\x5f\x48\x9c\x05\x31\x61\x6c\x68\xb1\x8d\x8f\xaa\x35\xef\x6d\xf5\xba\x44\xb0\xb0\x94\x9b\xd4\x86\x03\xbb\xf7\xa1\x96\x78\xff\x1c\x8c\xc6\x15\x34\x80\xcb\x83\xfd\xb0\x76\xaf\x79\xdc\x46\xd9\x5b\xb5\x3e\x01\x51\xbf\x50\x32\xa1\xd7\x04\x2e\x87\xd6\x3e\xd8\x4a\x2f\x55\x04\x36\x61\x8d\x65\x73\xb4\x65\xc2\x28\x37\x25\x4c\x61\x7d\x5f\x65\x62\x5f\x6d\x56\xca\x32\x0b\x75\x6d\x99\x3f\x34\x05\xb7\x11\x30\xe4\x6b\x9d\xc1\x62\x01\x51\xfb\x50\xa1\x98\xc0\xcc\x83\xaf\xdd\xc7\x53\x58\x80\xf7\x9b\xd0\x99\xc9\x8d\xad\xb3\xce\x65\x5f\x36\x56\xef\xce\x1e\xc2\xd6\x60\x62\x09\x13\xf8\x19\x86\xd6\x5c\xc1\x1a\xf5\x2b\xad\x25\x5d\x6d\x35\x0e\x07\xbd\xe5\x6e\x60\x63\x61\x30\x18\x55\xfd\xe4\xa0\xa8\x7f\x83\x71\x03\xf3\x08\x66\xee\xfd\x43\xf4\x67\x60\xca\x43\x9d\xcd\xbb\xba\x0a\xb5\xa4\xeb\x35\x4a\x13\x30\xef\x6c\x69\x3a\x15\xd2\xa5\xdf\xd9\xf1\xa9\x27\xb0\x77\x36\xac\xff\xf6\xfe\xed\xaf\xf6\x51\xd5\xa3\xdb\x3a\x91\x59\x0d\x28\xa7\xba\x38\xea\x2a\xe1\x60\x0c\x5a\x6e\x71\xec\x8c\xdb\x82\xb1\x1e\x63\x26\x11\xa2\xe3\xcc\x81\x59\x34\x8d\x63\x77\x55\xd0\xda\xb0\x65\xa8\x75\x19\xb5\xca\xe4\x7b\x09\xf6\x86\x48\x0b\xb8\xe1\x62\x0f\x7b\x04\x22\x11\xd6\xc2\x1e\x12\xb0\x21\x37\x08\x84\xc3\xab\x8f\xe4\x00\x12\xff\xbd\x45\xa5\x1b\x81\x46\x3e\x92\xc3\x1b\x46\xd6\xf7\x24\x33\x6b\xc0\xba\xc0\x25\x4c\xa0\x1a\x46\x35\x11\x39\x18\xc3\xc0\x15\xee\x47\x00\x18\xd3\x19\x00\xb3\xfd\x88\xe3\xb6\xf6\x0f\xc6\xb5\x9b\xa4\x42\x87\x24\xcf\x91\x27\xaf\x33\xca\x92\x61\x09\x3b\x6a\xea\xcd\x4c\x6d\x1b\xa2\xa9\x49\x40\xb7\xe0\x8c\x63\x95\x68\xc7\x05\x53\x1a\x28\xc3\xa2\xc1\xaa\xe0\x9c\xe7\x92\x24\xb1\xc6\xf9\x95\x2a\x8d\x1c\xe5\x70\x10\x67\x84\xaf\x0d\x27\xc7\xc1\xaa\xe9\x4b\x55\xe4\x0e\x31\xd0\x44\xae\x51\x07\x2e\x6e\x5f\xde\xe7\x9b\xc3\xda\xf6\x5d\x8b\xff\x63\xd5\x72\xad\x20\x4d\x1d\xf7\x45\x89\xa3\x0a\xc8\x8e\x50\x66\x2a\xcb\x91\xfd\x14\x86\xed\x92\xd6\xe4\xd3\x6a\xcf\x76\x0a\x46\x38\x23\xe9\x70\xd0\xf3\x61\x64\x30\x7a\xd9\xf6\xbd\xd7\xef\xde\x81\x19\x95\x2a\xc7\x3b\x32\x48\x15\xa8\x6d\x9e\x0b\xa9\x31\x81\xd5\xad\x3d\xbe\x92\x62\xaf\x50\x9e\xd5\x68\x7f\xb0\xb5\xd0\x98\xff\x58\xf9\x8a\x17\xe4\x49\xf9\x68\x06\xe9\xe3\xb2\x3e\xbe\x30\x24\x3b\x74\x2f\x22\x1f\xfc\x59\x56\x98\xda\xac\x6b\x43\xa9\x15\xe2\xce\x5d\xda\x06\xb5\x47\x4f\x5b\xb3\x30\x41\x2e\xed\xb9\x52\x03\x5b\xbe\x27\xdc\x8a\x88\x19\xd9\x51\xb1\x95\xaa\x05\x83\x41\x01\x72\x8d\x29\xd9\xb2\x86\x7d\xcb\x13\x4a\x8b\xfc\x77\x29\x72\xb2\x26\x2e\x27\xb5\x32\x6e\x33\x17\x34\xde\x3e\x9c\x50\xd0\xb7\xd5\x45\x57\x15\x7d\x4e\x53\xfb\xbe\xd8\xce\x8c\x0f\x8b\x50\x33\x65\xcd\xf6\xdf\xdc\xaa\x0f\x4a\xe2\xee\x56\xbe\x5a\x98\x7e\x2e\x9c\x24\xf7\x39\x57\xab\x0d\xc0\xc0\x14\xc7\xf7\xf6\x46\x1d\xa5\xcb\x1b\x36\xfc\x6c\xbc\xdb\x43\x3a\x23\x1a\xf6\x68\x5a\x5b\x07\xdb\xc0\x77\x4c\x3e\x75\xcc\x2d\x41\x7a\x93\xcf\x09\xd1\xee\x1a\x99\xa8\xcc\x3c\x26\x6f\xee\x89\x2a\x72\xa9\xae\x31\x71\x42\x11\x55\x45\x3c\xa5\x8a\x6e\x8c\x25\xdb\x9c\xd1\x98\x68\x74\x54\x94\xa2\x82\x57\xb9\xaf\xf8\xdd\x10\x10\xdc\xe6\x3f\xca\x21\x97\x62\x2d\x51\xd5\xe3\xd0\x76\x46\x4d\x33\x17\xb7\x47\xca\x1a\xba\x9a\x0d\x07\xa3\x51\xd9\xa6\xdb\x5a\xdd\xd0\xc6\x09\x97\xaf\x01\x77\x1c\xa1\xd7\xad\x6c\x17\x39\x68\x6a\xba\x3f\x4b\x9b\x92\x65\x4a\xb6\x2d\x4b\xc5\x3c\x65\xda\xf3\x8d\xbd\xe8\x2b\xf3\x69\x5d\xd2\xa6\x53\xf5\xe5\x9e\xc6\x81\x30\x84\x35\xd1\x19\xca\x52\xd7\xd6\xa0\xc6\xf5\x3a\x3d\x91\x61\xc4\x8c\x1b\xb0\x00\x8e\x7b\x28\xa7\x0f\xd7\xb7\xbf\xec\x34\xa2\x0d\xaf\xeb\xc4\xdc\xbd\x93\x41\x1d\x74\xdc\xec\x79\xbb\x98\xa0\x62\xac\x68\x00\x7a\x3b\x4d\xdb\x63\x8c\x5c\xff\xd8\xc9\xbf\x77\xed\xf0\xee\xe8\x88\xf4\xf5\x4d\x75\xc5\x14\x4a\xf9\x9f\xb7\xbf\xfe\x4d\xeb\xfc\x0f\x77\xb2\x9d\xc6\xcd\xc1\x40\xe4\xe8\x66\x9d\x16\x87\xee\xae\xd2\xf2\xd8\xdd\x74\x57\x97\x66\xd3\xf4\x91\x6d\x23\x3a\xbc\xfc\xbe\x59\xd5\xfd\x39\xed\x92\x27\x7c\x18\xec\xdd\x3a\x57\x82\x61\xc0\xc4\xda\xb6\x55\xa3\xd6\xbe\x31\xb6\xe5\x40\x69\xa2\xb7\x0a\x96\x0b\x98\x46\x91\x99\x44\xeb\xab\x73\xb8\x88\x7a\x66\x91\x62\x8a\x76\x6e\xf5\x5f\xef\xfe\xfe\x5b\x90\x13\xa9\xd0\x21\x94\xa8\x72\xc1\x15\xbe\xc7\x83\xee\xb0\xd5\x1b\x8e\x06\x51\x50\x7e\xee\x5e\x2c\xac\xb6\xe0\x67\x18\x1c\x7f\x75\x60\x00\x33\x68\xc4\x60\x1b\xab\x11\xe7\x49\x1d\xcf\xa8\x1a\x0d\x5b\x53\x8e\x3d\x64\xf7\x3a\x0e\xd5\x2e\xf0\x4c\x21\x10\x86\x52\x0f\x07\xf6\x1a\x27\x80\xdf\x19\x12\x85\x63\xfb\xfb\xab\x24\x76\x4d\xe8\x1e\x57\x1b\xa2\x34\xca\x27\x9d\x82\xd3\x6f\x71\x37\x8e\x7e\x7b\x93\xf7\xb3\xaa\xab\x8b\xa2\x47\xb1\xa7\x4c\x2c\x96\xb1\xd9\x2c\x98\x67\x2d\xd5\x84\xa1\xc9\xb6\x6c\x45\xe2\x1b\x37\xa2\xd8\x2f\xa1\x54\xf0\x7a\xce\x13\x2c\x41\xf9\x70\xca\xb3\xd7\x06\xf6\x1b\x7a\xf1\x85\x62\x50\x08\x6a\xd7\x06\xf0\x93\x0d\xd4\x6b\xa2\x71\x38\x32\x51\xf6\x9e\x6e\x70\xd8\x18\xc8\x9d\x17\x48\x07\x7d\x72\x1a\x72\xe8\xda\xa1\xf8\x7d\x05\xf8\xfd\x70\x30\x2f\x5e\xdc\xd5\xa5\x21\x5d\x63\xec\x27\x18\x78\x60\x3f\xeb\x2c\xbc\xe6\x77\x28\x6f\x39\x0f\xdd\xc1\x65\x07\xbf\x5b\xef\x1f\xa0\x8e\xc8\x3b\xe9\xd8\x01\x19\x62\x41\x41\xcb\x28\xc6\x50\x1b\xb4\x08\x54\xf2\xae\x44\x72\xdb\x18\xa9\x1c\x96\xbe\xe1\xb5\x35\xff\xd9\x49\xa7\xcd\x4f\x1f\x43\xdd\xe6\xc0\x8e\x19\xf7\x36\x6b\xfd\xe9\xa2\x40\x58\xfc\x32\xf8\x75\x43\x06\xca\x39\x4a\x33\xb2\x77\xdc\xfc\x51\xb1\xf1\x70\xdb\xfb\xc5\x49\xa7\x0f\xa5\x23\xde\x55\x63\x87\xe9\x6f\x91\x9f\x0a\x75\xe5\xc4\x74\x4e\xbf\x89\x04\x0b\xf2\xf7\x18\xfa\xee\x44\x04\xdb\x09\xb5\xde\x18\x96\x57\x61\x55\x33\x41\x53\xc8\x88\x02\x02\xa6\x12\xd8\xef\xce\x96\xa7\xb0\x60\xbf\x82\xbd\xb7\x25\x28\xb0\xd6\xbb\x46\xae\xe5\x6d\xd3\x3f\xec\x52\xdf\xa0\xce\x68\x7c\x73\x6f\xf3\xfd\xd0\x8c\xf6\x50\x2f\x37\x6e\x58\xbb\x1d\x7e\xb6\x1f\xb1\x4c\x0c\x4f\x0e\x0e\xed\x49\xff\x0d\x95\x98\x8a\x03\xb8\x6f\xc5\xab\xed\x1a\x52\x7a\xb0\x79\xd0\xf6\x82\x16\xe7\x83\x37\x14\x16\xf8\x74\x3c\x95\x8c\x75\x06\x7f\x07\xd7\xe2\xee\x21\x62\x2b\xb6\x95\x8f\xa7\x55\xaa\xf0\x24\xb9\xb3\x26\xe5\xbb\xde\xef\xc5\xad\x0f\xc6\xf3\xd0\x7d\x27\x3e\x9b\x87\xee\xff\xc2\xfc\x5f\x00\x00\x00\xff\xff\x68\x57\xc9\xa2\x1c\x33\x00\x00")

func ajaxuploadHtmlBytes() ([]byte, error) {
	return bindataRead(
		_ajaxuploadHtml,
		"ajaxupload.html",
	)
}

func ajaxuploadHtml() (*asset, error) {
	bytes, err := ajaxuploadHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "ajaxupload.html", size: 13084, mode: os.FileMode(420), modTime: time.Unix(1508812349, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _basicuploadHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x90\x41\x6a\x03\x31\x0c\x45\xf7\x3e\x85\xf0\xbe\x9d\x0b\x78\x66\x99\x65\x29\xb4\x3d\x80\x12\x2b\xd8\xd4\xb2\xcc\x44\x0e\x84\x90\xbb\x17\x8d\xdb\x32\x8b\x76\xe9\xcf\xff\x0f\xeb\x85\xa4\x5c\x16\xe7\x42\x22\x8c\x8b\x03\x08\x9a\xb5\xd0\xf2\xd1\x8a\x60\x84\x73\x2e\x14\xa6\x11\xb9\x30\x8d\x92\x0b\x47\x89\xb7\xad\x7c\x96\x95\x81\xea\x49\x6f\x8d\x66\xcf\xbd\x68\x6e\xb8\xea\x64\xf9\x53\x44\x45\x0f\x78\xd2\x2c\x75\xf6\xf7\xfb\xf3\x80\xbe\xd1\x7a\xa5\x57\xd4\xf4\x78\x78\x60\xd2\x24\x71\xf6\x4d\x2e\xea\x0d\x09\x10\x72\x6d\x5d\x61\x20\xed\x03\x1e\x2a\x32\xed\x08\x87\x5c\xe8\x20\x2b\xbf\x20\x93\x41\xa6\x3f\x86\x29\xc7\x48\xf5\x67\xaa\xf2\x69\x8f\x2b\x96\xbe\x07\xbd\x5b\xfc\x1f\xe1\xd2\x8f\x9c\xf5\x77\xd4\xb7\xc5\x77\x35\x6c\x17\x9a\x92\xa1\xc2\xdc\x98\xc8\xaf\x00\x00\x00\xff\xff\x3e\x84\x3a\x47\x4f\x01\x00\x00")

func basicuploadHtmlBytes() ([]byte, error) {
	return bindataRead(
		_basicuploadHtml,
		"basicupload.html",
	)
}

func basicuploadHtml() (*asset, error) {
	bytes, err := basicuploadHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "basicupload.html", size: 335, mode: os.FileMode(420), modTime: time.Unix(1508812349, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _multipleuploadHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x93\xcf\x6a\xdc\x30\x10\xc6\xef\x7a\x8a\xa9\xee\x5e\x15\x7a\x0b\xb2\xa0\x87\x2c\x0d\x24\x6d\x60\xb7\x50\x7a\x09\xf2\x6a\x76\xad\x56\x7f\x8c\x3c\x36\x0d\x21\xef\x5e\x24\xd9\x1b\x13\xb2\xb7\x99\xfd\xe6\x37\xdf\x8c\xc6\xb2\x27\xef\x14\x93\x9f\x9a\x06\x7a\xa2\xe1\x46\x88\xce\xc5\xcb\x8e\x50\x7b\x4a\x88\x7d\x9c\x46\xdc\x9d\xa2\x17\xd3\xe0\xa2\x36\x36\x5c\x9a\xb3\x75\x38\x36\xfa\x8f\xfe\x07\x4d\xa3\x98\xec\x51\x1b\xc5\x00\x24\x59\x72\xa8\x7e\x16\x21\x64\x95\x14\x35\xc5\xa4\xa8\x22\x26\xbb\x68\x9e\x15\xcb\xea\x73\x4c\x1e\xac\x69\x79\x25\x73\xd0\x27\xb2\x31\xb4\xfc\xe5\x65\x57\x19\x07\x4c\x33\x3e\x6a\xea\x5f\x5f\x39\x78\xa4\x3e\x9a\x96\x3f\xfe\x38\x1c\x39\x60\x38\xd1\xf3\x80\x2d\xf7\x93\x23\x3b\xe8\x44\x22\xf3\x1a\xa3\x49\xf3\xc2\x07\x90\x36\x0c\x13\x41\xd5\xf5\xd6\x18\x0c\x1c\x82\xf6\xd8\x72\x8a\x7f\x73\x30\x6b\x37\xe1\xa6\xe1\x31\xa7\x73\x33\xa1\x2a\xe1\x6c\xd1\x99\x11\xa9\x86\x00\xd2\xe1\x05\x83\x51\xdf\x8e\x0f\xf7\xb0\xb7\x0e\xa1\x16\x4a\xb1\xfc\xc1\x56\xe1\x47\xbd\xf3\xb0\x0f\x5f\x7f\x3d\xed\xef\xee\x6f\x9f\x0e\x77\xbf\x6f\x57\x3b\xef\x92\x8b\xad\x2f\x9f\xf3\xaf\x98\x59\xa9\xc6\xce\xab\x95\x6c\x46\x77\xe8\xe0\x1c\x53\xcb\xcb\x9b\xa0\xc3\x13\x71\x95\x7d\x8d\x40\x11\xea\x62\x6f\xa4\x28\xc2\x4d\xe1\xd6\x5c\xae\xac\xd6\x36\x8c\xc5\xd7\x75\x31\x19\xb9\x8f\xc9\x7f\xd7\x1e\xcb\x63\x94\xad\xbb\xeb\xfe\x33\x42\x6c\x1a\x18\x3b\x5f\x91\x26\xe9\x0b\x57\x31\x81\x49\x71\x28\x67\x31\x42\x8f\x09\xa5\xd8\x4c\xb3\x04\xec\x5d\xfd\x38\x75\xde\x52\x37\x11\xc5\xc0\x37\xfc\x9a\x59\x26\xa8\x22\xbe\x1e\x5e\x19\x5f\x8a\x2a\xf9\x88\x2f\xc5\xdb\xb3\xb2\x12\xc6\xe4\xf3\x8d\x2e\xb7\x29\x45\xf9\x28\xfe\x07\x00\x00\xff\xff\x22\x8f\x0a\x70\x1b\x03\x00\x00")

func multipleuploadHtmlBytes() ([]byte, error) {
	return bindataRead(
		_multipleuploadHtml,
		"multipleupload.html",
	)
}

func multipleuploadHtml() (*asset, error) {
	bytes, err := multipleuploadHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "multipleupload.html", size: 795, mode: os.FileMode(420), modTime: time.Unix(1508812349, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"BUILD.bazel": buildBazel,
	"ajaxupload.html": ajaxuploadHtml,
	"basicupload.html": basicuploadHtml,
	"multipleupload.html": multipleuploadHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"BUILD.bazel": &bintree{buildBazel, map[string]*bintree{}},
	"ajaxupload.html": &bintree{ajaxuploadHtml, map[string]*bintree{}},
	"basicupload.html": &bintree{basicuploadHtml, map[string]*bintree{}},
	"multipleupload.html": &bintree{multipleuploadHtml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}


func assetFS() *assetfs.AssetFS {
	assetInfo := func(path string) (os.FileInfo, error) {
		return os.Stat(path)
	}
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: assetInfo, Prefix: k}
	}
	panic("unreachable")
}
