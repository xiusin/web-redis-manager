package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/xiusin/logger"
	"github.com/xiusin/redis_manager/server/src"
)

var DEBUG = false

var once sync.Once

var cacheDir string

var mux *http.ServeMux

const SecretLen = 32

var secretKey []byte

func init() {
	cacheDir = GetCacheDir()

	var routes = map[string]src.HandleFunc{
		"/redis/connection/test":        src.RedisManagerConnectionTest,
		"/redis/connection/save":        src.RedisManagerConfigSave,
		"/redis/connection/list":        src.RedisManagerConnectionList,
		"/redis/connection/server":      src.RedisManagerConnectionServer,
		"/redis/connection/removekey":   src.RedisManagerRemoveKey,
		"/redis/connection/removerow":   src.RedisManagerRemoveRow,
		"/redis/connection/updatekey":   src.RedisManagerUpdateKey,
		"/redis/connection/addkey":      src.RedisManagerAddKey,
		"/redis/connection/flushDB":     src.RedisManagerFlushDB,
		"/redis/connection/remove":      src.RedisManagerRemoveConnection,
		"/redis/connection/command":     src.RedisManagerCommand,
		"/redis/connection/info":        src.RedisManagerGetInfo,
		"/redis/connection/get-command": src.RedisManagerGetCommandList,
	}

	mux = http.NewServeMux()

	for route, handle := range routes {
		mux.HandleFunc(route, func(handle src.HandleFunc) func(writer http.ResponseWriter, request *http.Request) {
			return func(writer http.ResponseWriter, request *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						logger.Errorf("Recovered Error: %s", err)
					}
				}()
				var params url.Values
				data := make(map[string]interface{})
				if request.Method == http.MethodPost {
					request.ParseForm()
					params = request.PostForm
				} else {
					params = request.URL.Query()
				}
				for param, values := range params {
					if len(values) > 0 {
						data[param] = values[0]
					} else {
						data[param] = nil
					}
				}
				writer.Write([]byte(handle(data)))
			}
		}(handle))
	}
}

//go:embed resources
var embededFiles embed.FS

func getFileSystem(useOS bool) http.FileSystem {
	fsys, err := fs.Sub(embededFiles, "resources/app")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

func main() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  10240,
		WriteBufferSize: 10240,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: true,
	}
	mux.HandleFunc("/redis/connection/pubsub", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			data := make(map[string]interface{})
			request.ParseForm()
			params := request.PostForm
			for param, values := range params {
				if len(values) > 0 {
					data[param] = values[0]
				} else {
					data[param] = nil
				}
			}
			writer.Write([]byte(src.RedisPubSub(data)))
			return
		}

		ws, _ := upgrader.Upgrade(writer, request, nil)
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				continue
			}
			data := make(map[string]interface{})
			if err := json.Unmarshal(msg, &data); err != nil {
				continue
			}
			data["ws"] = ws
			src.RedisPubSub(data)
		}
	})

	mux.Handle("/", http.FileServer(getFileSystem(true)))

	handler := cors.Default().Handler(mux)
	fmt.Println("start rdm server in http://0.0.0.0:8787")
	_ = http.ListenAndServe(":8787", handler)
}
func GetCacheDir() string {
	once.Do(func() {
		var workingDir string
		if DEBUG {
			workingDir, _ = os.Getwd()
		} else {
			workingDir, _ = os.Executable()
			workingDir = filepath.Dir(workingDir)
		}
		cacheDir = workingDir
		src.ConnectionFile = fmt.Sprintf("%s/data.db", cacheDir)
	})
	return cacheDir
}

func gek(_ src.RequestData) string {
	return src.JSON(src.ResponseData{
		Status: src.SuccessCode,
		Msg:    "success",
		Data:   src.SecretKey,
	})
}

func getRandomKey() {
	salt := "ABCEDFGHIJKLMNOPQRSTUVWXYZ0123456789"
	l := len(salt)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < SecretLen; i++ {
		secretKey = append(secretKey, salt[rand.Int63n(int64(l))])
	}
	src.SecretKey = string(secretKey)
}
