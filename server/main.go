package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/xiusin/redis_manager/server/handler"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xiusin/redis_manager/server/windows"

	"github.com/kataras/basicauth"
	"github.com/rs/cors"
	"github.com/xiusin/redis_manager/server/router"
)

var (
	mux           = http.NewServeMux()
	basicAuthName string
	basicAuthPass string
	port          = ":8787"
)

//go:embed resources
var embedFiles embed.FS

func init() {
	handler.ConnectionFile = windows.GetStorePath("rdm.db")
	router.RegisterRouter(mux)
}

func main() {
	isDebug := strings.Contains(os.Args[0], "build")
	flag.StringVar(&basicAuthName, "username", "admin", "basicAuth 名称")
	flag.StringVar(&basicAuthPass, "password", "", "basicAuth 验证密码")
	flag.Parse()

	appAssets, err := fs.Sub(embedFiles, "resources/app")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", http.FileServer(http.FS(appAssets)))

	_handler := cors.Default().Handler(mux)
	if len(basicAuthName) > 0 && len(basicAuthPass) > 0 {
		_handler = basicauth.Default(map[string]string{basicAuthName: basicAuthPass})(_handler)
	}

	if !isDebug {
		go func() { _ = http.ListenAndServe(port, _handler) }()

		time.Sleep(time.Millisecond * 100)

		portInt, _ := strconv.Atoi(strings.Trim(port, ":"))
		windows.InitWebview(fmt.Sprintf("http://localhost:%d/#/", portInt))
	} else {
		_ = http.ListenAndServe(port, _handler)
	}

}
