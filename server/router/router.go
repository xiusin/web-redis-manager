package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/xiusin/redis_manager/server/src"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  102400,
	WriteBufferSize: 102400,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func RegisterRouter(mux *http.ServeMux) {

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
		"/redis/connection/renameKey":   src.RedisManagerRenameKey,
		"/redis/connection/command":     src.RedisManagerCommand,
		"/redis/connection/info":        src.RedisManagerGetInfo,
		"/redis/connection/get-command": src.RedisManagerGetCommandList,
	}

	notReadonlyKeys := []string{"removekey", "removerow", "updatekey", "addkey", "flushDB", "renameKey", "command"}

	for route, handle := range routes {
		mux.HandleFunc(route, func(handle src.HandleFunc) func(writer http.ResponseWriter, request *http.Request) {
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				defer func() {
					if err := recover(); err != nil {
						writer.Write([]byte(src.JSON(src.ResponseData{Status: src.FailedCode, Msg: err.(error).Error()})))
					}
				}()

				// 检查是否为只读模式
				modifyKey := strings.Replace(request.URL.Path, "/redis/connection/", "", 1)
				isReadonly := false

				if isReadonly {
					for _, key := range notReadonlyKeys {
						if modifyKey == key {
							writer.Write([]byte(src.JSON(src.ResponseData{Status: src.FailedCode, Msg: "只读模式下不可做修改或新增操作", Data: nil})))
							return
						}
					}
				}

				var params url.Values
				data := make(map[string]interface{})
				if request.Method == http.MethodPost {
					_ = request.ParseForm()
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

	mux.HandleFunc("/redis/connection/pubsub", func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recover", err)
			}
		}()
		if request.Method == http.MethodPost {
			data := make(map[string]interface{})
			_ = request.ParseForm()
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
}
