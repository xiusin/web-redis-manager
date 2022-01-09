package main

import (
	"embed"
	"fmt"
	"github.com/kataras/basicauth"
	"github.com/rs/cors"
	"github.com/xiusin/redis_manager/server/router"
	"github.com/xiusin/redis_manager/server/src"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

var cacheDir string
var mux = http.NewServeMux()

//go:embed resources
var embedFiles embed.FS

var port = ":8787"

func init() {
	cacheDir, _ = os.Getwd()
	src.ConnectionFile = filepath.Join(cacheDir, "data.db")
	router.RegisterRouter(mux)
}

func main() {
	fsys, err := fs.Sub(embedFiles, "resources/app")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", http.FileServer(http.FS(fsys)))
	handler := cors.Default().Handler(mux)
	if basicauthPass := os.Getenv("RDM_PASS"); len(basicauthPass) > 0 {
		basicauthName := "admin"
		auth := basicauth.Default(map[string]string{
			basicauthName: basicauthPass,
		})
		handler = auth(handler)
	}

	fmt.Println("start rdm server in http://0.0.0.0" + port)
	_ = http.ListenAndServe(port, handler)
}
