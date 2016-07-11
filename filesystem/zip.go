package filesystem

/*
import (
	"net/http"
	"archive/zip"
)

func newZipFs(filePath string) (http.FileSystem, error) {
    fs := &zipFs{
        filePath: filePath,
    }
    err := fs.indexFs()
    if err != nil {
        return nil, err
    }
    return fs, nil
}

type zipFs struct {
    filePath string
}

func (this *zipFs) indexFs() error {
    r, err := zip.OpenReader(this.filePath)
    if err != nil {
        return err
    }
    defer r.Close()
    // TODO(jeremyje): Clone the file in memory to allow for closing of the zip file.
    for _, f := range r.File {
        if name == f.Name {
            return newZipFile(f)
        }
    }
}

func newZipFile(fp *zip.File) *zipFile {

}

type zipFile struct {

}

func (this *zipFile) Close() error {
    return nil
}

func (this *zipFile) Read(p []byte) (n int, err error) {

}

func (this *zipFile) Seek(offset int64, whence int) (int64, error) {

}

func (this *zipFile) Readdir(count int) ([]os.FileInfo, error) {

}

func (this *zipFile) Stat() (os.FileInfo, error) {

}


func (this *zipFs) Open(name string) (File, error) {
    r, err := zip.OpenReader(this.filePath)
    if err != nil {
        return err
    }
    defer r.Close()
    // TODO(jeremyje): Clone the file in memory to allow for closing of the zip file.
    for _, f := range r.File {
        if name == f.Name {
            return f.Open()
        }
    }
}
*/
