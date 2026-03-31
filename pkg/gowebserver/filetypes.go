package gowebserver

import (
	"log"
	"mime"
	"path/filepath"
	"strings"
)

var (
	mimeIconMap = map[string]string{
		".":                               "folder",
		"image":                           "image",
		"application/pdf":                 "pdf",
		"audio":                           "audio",
		"text":                            "text",
		"video":                           "video",
		".txt":                            "text",
		".pdf":                            "pdf",
		".doc":                            "doc",
		".xls":                            "spreadsheet",
		".ppt":                            "presentation",
		".jpg":                            "image",
		".mp4":                            "video",
		".xvid":                           "video",
		".mp3":                            "audio",
		".zip":                            "archive",
		".cc":                             "code",
		".go":                             "code",
		".cs":                             "code",
		".java":                           "code",
		".cpp":                            "code",
		".sh":                             "terminal",
		".rar":                            "archive",
		".7z":                             "archive",
		".xz":                             "archive",
		".bz2":                            "archive",
		".tar":                            "archive",
		".gz":                             "archive",
		".ps1":                            "terminal",
		".psm1":                           "terminal",
		".cmd":                            "terminal",
		".bash":                           "terminal",
		".download":                       "download",
		".exe":                            "terminal",
		"application/x-shellscript":       "terminal",
		"application/x-ms-dos-executable": "terminal",
		"application/x-msdownload":        "terminal",
		".db":                             "database",
		".epub":                           "ebook",
		".dwg":                            "cad",
		".svg":                            "vector",
		".psd":                            "photoshop",
		".html":                           "markup",
		".htm":                            "markup",
		".css":                            "stylesheet",
		".scss":                           "stylesheet",
		".js":                             "script",
		".ts":                             "script",
		".tsx":                            "script",
		".dat":                            "data",
		".crt":                            "certificate",
		".cert":                           "certificate",
		".pem":                            "key",
		".pkv":                            "key",
		".pk":                             "key",
		".key":                            "key",
		".log":                            "log",
		".bak":                            "backup",
		".bin":                            "binary",
		".pkg":                            "package",
		".rpm":                            "package",
		".msi":                            "package",
		".deb":                            "package",
		".snap":                           "package",
		".sqlite":                         "database",
		".pub":                            "certificate",
		"application/x-x509-ca-cert":      "certificate",
		"application/x-yaml":              "config",
		"application/illustrator":         "photoshop",
		".ds_store":                       "database",
		".ini":                            "config",
		"application/json":                "config",
		"font":                            "font",
		".config":                         "config",
		".cfg":                            "config",
		".yaml":                           "config",
		".yml":                            "config",
		"application/x-cd-image":          "disc",
		".iso":                            "disc",
		".docx":                           "doc",
		".xlsx":                           "spreadsheet",
		".pptx":                           "presentation",
		".md":                             "doc",
		".ttf":                            "font",
		".ai":                             "photoshop",
		".webm":                           "video",
	}
)

func nameToIconClass(isDir bool, name string) string {
	ext := filepath.Ext(strings.ToLower(name))
	if isDir {
		return "folder"
	}

	if val, ok := mimeIconMap[ext]; ok {
		return val
	}

	mimeType := mime.TypeByExtension(ext)

	if mimeType != "" {
		if val, ok := mimeIconMap[mimeType]; ok {
			return val
		}

		if parts := strings.Split(mimeType, "/"); len(parts) > 1 {
			if val, ok := mimeIconMap[parts[0]]; ok {
				return val
			}
		}
	}

	log.Printf("%s > %s", ext, mimeType)
	return "unknown"
}
