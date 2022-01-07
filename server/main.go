package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/cors"
	"github.com/xiusin/redis_manager/server/router"
	"github.com/xiusin/redis_manager/server/src"
)

var cacheDir string
var mux = http.NewServeMux()

//go:embed resources
var embededFiles embed.FS

var port = ":8787"

func getFileSystem(useOS bool) http.FileSystem {
	fsys, err := fs.Sub(embededFiles, "resources/app")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

func init() {
	cacheDir, _ = os.Getwd()
	src.ConnectionFile = filepath.Join(cacheDir, "data.db")
	router.RegisterRouter(mux)
}

func main() {
	mux.Handle("/", http.FileServer(getFileSystem(true)))
	handler := cors.Default().Handler(mux)

	fmt.Println("start rdm server in http://0.0.0.0" + port)
	_ = http.ListenAndServe(port, handler)
}
