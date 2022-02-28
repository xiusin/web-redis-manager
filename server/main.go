package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xiusin/redis_manager/server/windows"

	"github.com/kataras/basicauth"
	"github.com/rs/cors"
	"github.com/xiusin/redis_manager/server/router"
	"github.com/xiusin/redis_manager/server/src"
)

var mux = http.NewServeMux()
var basicauthName string
var basicauthPass string

//go:embed resources
var embedFiles embed.FS

var port = ":8787"

func init() {
	src.ConnectionFile = windows.GetStorePath("rdm.db")
	router.RegisterRouter(mux)
}

func main() {
	isDebug := strings.Contains(os.Args[0], "build")

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

	if !isDebug {
		go func() {
			_ = http.ListenAndServe(port, handler)
		}()

		time.Sleep(time.Millisecond * 100)
		windows.InitWebview(strings.Trim(port, ":"), !isDebug)
	} else {
		fmt.Println("start rdm server in http://0.0.0.0" + port)
		fmt.Println(http.ListenAndServe(port, handler))
	}

}
