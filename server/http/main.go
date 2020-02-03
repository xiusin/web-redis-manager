package main

import (
  "encoding/json"
  "fmt"
  "github.com/asticode/go-astilog"
  "github.com/gorilla/websocket"
  "github.com/rs/cors"
  "github.com/xiusin/redis_manager/server/src"
  "net/http"
  "net/url"
  "os"
  "path/filepath"
  "runtime/debug"
  "sync"
)

const DEBUG = true

var once sync.Once

var cacheDir string

var mux *http.ServeMux

func init() {
  cacheDir = GetCacheDir()

  astilog.SetLogger(astilog.New(astilog.Configuration{
    AppName:  "RedisManager",
    Filename: fmt.Sprintf("%s/rdm-log.log", cacheDir),
    Verbose:  DEBUG,
  }))
  astilog.FlagConfig()

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
            s := debug.Stack()
            astilog.Errorf("Recovered Error: %s, ErrorStack: \n%s\n\n", err, string(s))
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
      _, msg, _ := ws.ReadMessage()
      data := make(map[string]interface{})
      if err := json.Unmarshal(msg, &data); err != nil {
        astilog.GetLogger().Error(err)
        continue
      }
      data["ws"] = ws
      src.RedisPubSub(data)
    }
  })

  handler := cors.Default().Handler(mux)

  http.ListenAndServe(":18998", handler)
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
