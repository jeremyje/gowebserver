package server

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html
import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/jeremyje/gowebserver/embedded"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

const (
	uploadHtmlPage     = "ajaxupload.html"
	uploadFileFormName = "gowebserveruploadfile[]"
)

type uploadHttpHandler struct {
	uploadServePath string
	uploadDirectory string
}

type uploadResponse struct {
	Success bool `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (this *uploadHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		tmpl := template.New("")
		t, err := tmpl.Parse(string(embedded.MustAsset(uploadHtmlPage)))
		if err != nil {
			fmt.Printf("Error parsing html template: %s", err)
			w.Write([]byte(err.Error()))
			return
		}
		var params = struct {
			UploadServePath    string
			UploadToken        string
			UploadFileFormName string
		}{this.uploadServePath, token, uploadFileFormName}
		t.Execute(w, params)
	} else {
		var resp uploadResponse
		r.ParseMultipartForm(32 << 20)
		//file, handler, err := r.FormFile(uploadFileFormName)
		m := r.MultipartForm
		files := m.File[uploadFileFormName]
		for i, _ := range files {
			file, err := files[i].Open()
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot download file (%s), %s", files[i].Filename, err)
				writeUploadResponse(w, resp)
				return
			}
			defer file.Close()
			err = os.MkdirAll(this.uploadDirectory, 0766)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create directory to store file (%s), %s", this.uploadDirectory, err)
				writeUploadResponse(w, resp)
				return
			}
			localPath := filepath.Join(this.uploadDirectory, files[i].Filename)
			f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create file (%s), %s", localPath, err)
				writeUploadResponse(w, resp)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		}
		
		resp.Success = true
		writeUploadResponse(w, resp)
	}
}

func (this *uploadHttpHandler) ServeHTTPOld(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		tmpl := template.New("")
		t, err := tmpl.Parse(string(embedded.MustAsset(uploadHtmlPage)))
		if err != nil {
			fmt.Printf("Error parsing html template: %s", err)
			w.Write([]byte(err.Error()))
			return
		}
		var params = struct {
			UploadServePath    string
			UploadToken        string
			UploadFileFormName string
		}{this.uploadServePath, token, uploadFileFormName}
		t.Execute(w, params)
	} else {
		var resp uploadResponse
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile(uploadFileFormName)
		if err != nil {
			resp.Error = fmt.Sprintf("InternalError: The form tag for file (%s), %s", uploadFileFormName, err)
			writeUploadResponse(w, resp)
			return
		}
		defer file.Close()
		err = os.MkdirAll(this.uploadDirectory, 0766)
		if err != nil {
			resp.Error = fmt.Sprintf("InternalError: Cannot create directory to store file (%s), %s", this.uploadDirectory, err)
			writeUploadResponse(w, resp)
			return
		}
		localPath := filepath.Join(this.uploadDirectory, handler.Filename)
		f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			resp.Error = fmt.Sprintf("InternalError: Cannot create file (%s), %s", localPath, err)
			writeUploadResponse(w, resp)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		resp.Success = true
		writeUploadResponse(w, resp)
	}
}

func newUploadHandler(uploadServePath string, uploadDirectory string) http.Handler {
	return &uploadHttpHandler{
		uploadServePath: uploadServePath,
		uploadDirectory: uploadDirectory,
	}
}

func writeUploadResponse(w http.ResponseWriter, resp uploadResponse) {
	if len(resp.Error) > 0 {
		log.Println(resp.Error)
		w.WriteHeader(400)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "InternalError: Malformed server response.", 500)
		log.Printf("Error: Cannot marshal upload JSON response, %s", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
