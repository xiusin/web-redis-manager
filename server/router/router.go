package router

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/xiusin/logger"
	"github.com/xiusin/rdm/server/handler"
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

	for route, handle := range routes {
		mux.HandleFunc(route, func(handle handler.HandleFunc) func(writer http.ResponseWriter, request *http.Request) {
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				defer func() {
					if err := recover(); err != nil {
						_, _ = writer.Write([]byte(handler.JSON(handler.ResponseData{Status: handler.FailedCode, Msg: err.(error).Error()})))
					}
				}()

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
				if data["id"] != nil {
					cfg, err := handler.GetServerCfg(data)
					handler.ThrowIf(err)
					handler.CheckReadonly(cfg.Readonly, strings.Replace(request.URL.Path, "/redis/connection/", "", 1))
				}
				_, _ = writer.Write([]byte(handle(data)))
			}
		}(handle))
	}

	mux.HandleFunc("/redis/connection/pubsub", func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
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
				break
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
