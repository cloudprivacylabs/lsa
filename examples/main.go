package examples

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed contact/*
var EmbededFiles embed.FS

func GetFileSystem(useOS bool) http.FileSystem {
	if useOS {
		log.Print("using live mode")
		return http.FS(os.DirFS("static"))
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(EmbededFiles, "static")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
