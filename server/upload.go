package server

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html
const (
	uploadPage = `<html>
<head>
       <title>Upload file</title>
</head>
<body>
<form enctype="multipart/form-data" action="{{.UploadServePath}}" method="post">
    <input type="file" name="uploadfile" />
    <input type="hidden" name="token" value="{{.UploadToken}}"/>
    <input type="submit" value="upload" />
</form>
</body>
</html>
`
)

type uploadHttpHandler struct {
	uploadServePath string
	uploadDirectory string
}

func (this *uploadHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		tmpl := template.New("")
		t, err := tmpl.Parse(uploadPage)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte(err.Error()))
			return
		}
		var params = struct {
			UploadServePath string
			UploadToken     string
		}{this.uploadServePath, token}
		t.Execute(w, params)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		localPath := filepath.Join(this.uploadDirectory, handler.Filename)
		f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func newUploadHandler(uploadServePath string, uploadDirectory string) http.Handler {
	return &uploadHttpHandler{
		uploadServePath: uploadServePath,
		uploadDirectory: uploadDirectory,
	}
}
