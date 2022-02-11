package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/xiusin/redis_manager/server/windows"

	"github.com/kataras/basicauth"
	"github.com/rs/cors"
	"github.com/xiusin/redis_manager/server/router"
	"github.com/xiusin/redis_manager/server/src"
)

var cacheDir string
var mux = http.NewServeMux()
var basicauthName string
var basicauthPass string

//go:embed resources
var embedFiles embed.FS

var port = ":8787"

var IsBuildStr string

func init() {
	src.ConnectionFile = windows.GetStorePath("rdm_" + runtime.GOOS + ".db")
	router.RegisterRouter(mux)
	IsBuildStr = "true"
}

func main() {

	flag.StringVar(&basicauthName, "username", "admin", "basicauth 名称")
	flag.StringVar(&basicauthPass, "password", "", "basicauth 验证密码")

	flag.Parse()

	fsys, err := fs.Sub(embedFiles, "resources/app")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", http.FileServer(http.FS(fsys)))
	handler := cors.Default().Handler(mux)
	if len(basicauthName) > 0 && len(basicauthPass) > 0 {
		auth := basicauth.Default(map[string]string{
			basicauthName: basicauthPass,
		})
		handler = auth(handler)
	}

	go func() {
		fmt.Println("start rdm server in http://0.0.0.0" + port)
		_ = http.ListenAndServe(port, handler)
	}()

	time.Sleep(time.Millisecond * 100)
	isBuild, _ := strconv.ParseBool(IsBuildStr)
	windows.InitWebview(strings.Trim(port, ":"), isBuild)
}
