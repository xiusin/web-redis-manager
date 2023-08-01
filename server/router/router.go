package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/xiusin/logger"
	"github.com/xiusin/redis_manager/server/handler"
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
	var routes = map[string]handler.HandleFunc{
		"/redis/connection/test":        handler.RedisManagerConnectionTest,
		"/redis/connection/save":        handler.RedisManagerConfigSave,
		"/redis/connection/list":        handler.RedisManagerConnectionList,
		"/redis/connection/server":      handler.RedisManagerConnectionServer,
		"/redis/connection/removekey":   handler.RedisManagerRemoveKey,
		"/redis/connection/removerow":   handler.RedisManagerRemoveRow,
		"/redis/connection/updatekey":   handler.RedisManagerUpdateKey,
		"/redis/connection/addkey":      handler.RedisManagerAddKey,
		"/redis/connection/flushDB":     handler.RedisManagerFlushDB,
		"/redis/connection/remove":      handler.RedisManagerRemoveConnection,
		"/redis/connection/renameKey":   handler.RedisManagerRenameKey,
		"/redis/connection/command":     handler.RedisManagerCommand,
		"/redis/connection/info":        handler.RedisManagerGetInfo,
		"/redis/connection/get-command": handler.RedisManagerGetCommandList,
	}

	notReadonlyKeys := []string{"removekey", "removerow", "updatekey", "addkey", "flushDB", "renameKey", "command"}

	for route, handle := range routes {
		mux.HandleFunc(route, func(handle handler.HandleFunc) func(writer http.ResponseWriter, request *http.Request) {
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				defer func() {
					if err := recover(); err != nil {
						logger.Print(string(debug.Stack()))
						_, _ = writer.Write([]byte(handler.JSON(handler.ResponseData{Status: handler.FailedCode, Msg: err.(error).Error()})))
					}
				}()

				// 检查是否为只读模式
				modifyKey := strings.Replace(request.URL.Path, "/redis/connection/", "", 1)
				isReadonly := false

				if isReadonly {
					for _, key := range notReadonlyKeys {
						if modifyKey == key {
							_, _ = writer.Write([]byte(handler.JSON(handler.ResponseData{Status: handler.FailedCode, Msg: "只读模式下不可做修改或新增操作", Data: nil})))
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

				_, _ = writer.Write([]byte(handle(data)))
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
			_, _ = writer.Write([]byte(handler.RedisPubSub(data)))
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
			handler.RedisPubSub(data)
		}
	})
}
