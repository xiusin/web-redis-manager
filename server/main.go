package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/basicauth"
	"github.com/rs/cors"
	"github.com/xiusin/rdm/server/handler"
	"github.com/xiusin/rdm/server/router"
	"github.com/xiusin/rdm/server/windows"
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
	var hasAuth = len(basicAuthName) > 0 && len(basicAuthPass) > 0
	if hasAuth {
		_handler = basicauth.Default(map[string]string{basicAuthName: basicAuthPass})(_handler)
	}

	if !isDebug {
		go func() { _ = http.ListenAndServe(port, _handler) }()
		time.Sleep(time.Millisecond * 100)
		portInt, _ := strconv.Atoi(strings.Trim(port, ":"))
		windows.InitWebview(fmt.Sprintf("http://localhost:%d/#/", portInt))
	} else {
		fmt.Printf("> service listening on: \033[32mhttp://0.0.0.0:%s/ \033[0m\n", strings.Trim(port, ":"))
		fmt.Println("> \033[43;35missue: https://github.com/xiusin/web-redis-manager/issues\033[0m")
		if hasAuth {
			fmt.Println(`> 账号: ` + basicAuthName + ` 密码: ` + basicAuthPass)
		}
		_ = http.ListenAndServe(port, _handler)
	}

}
