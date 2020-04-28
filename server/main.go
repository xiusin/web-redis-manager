package main

import (
  "encoding/json"
  "fmt"
  "github.com/Luzifer/go-openssl"
  "github.com/asticode/go-astilectron"
  bootstrap "github.com/asticode/go-astilectron-bootstrap"
  "github.com/asticode/go-astilog"
  "github.com/pkg/errors"
  "github.com/xiusin/logger"
  "github.com/xiusin/redis_manager/server/src"
  "log"
  "math/rand"
  "os"
  "path/filepath"
  "strings"
  "sync"
  "time"
)

var DEBUG = false

var once sync.Once

var cacheDir string

var handler = src.NewHandler()

const SecretLen = 32

var secretKey []byte

func init() {
  cacheDir = GetCacheDir()
  getRandomKey()
  f,err := os.OpenFile(fmt.Sprintf("%s/error.log", cacheDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
  if err == nil {
    logger.SetOutput(f)
  }
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
    "/redis/connection/pubsub":      src.RedisPubSub,
    "/redis/connection/info":        src.RedisManagerGetInfo,
    "/gek":                          gek,
    "/redis/connection/get-command": src.RedisManagerGetCommandList,
  }

  for route, handle := range routes {
    handler.Add(route, handle)
  }
}

func main() {
  options := astilectron.Options{
    AppName:            "RedisManager",
    SingleInstance:     true,
    BaseDirectoryPath:  cacheDir,
    AppIconDefaultPath: fmt.Sprintf("%s/resources/icon.png", cacheDir),
    DataDirectoryPath:  cacheDir,
  }

  var url string
  if DEBUG {
    url = "http://localhost:8899"
  } else {
    url = "index.html"
  }
  center, HasShadow, FullScreenable, Closable, skipTaskBar := true, true, true, true, true
  height, width := 800, 1280

  config := bootstrap.Options{
    Asset:              Asset,
    AssetDir:           AssetDir,
    RestoreAssets:      RestoreAssets,
    AstilectronOptions: options,
    Debug:              DEBUG,
    MenuOptions: nil,
    OnWait: func(a *astilectron.Astilectron, ws []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
      a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
        log.Println("App has crashed")
        return
      })
      a.On(astilectron.EventNameAppClose, func(e astilectron.Event) (deleteListener bool) {
        fmt.Println("astilectron.EventNameAppClose")
        return
      })
      a.On(astilectron.EventNameAppCmdQuit, func(e astilectron.Event) (deleteListener bool) {
        fmt.Println("astilectron.EventNameAppCmdQuit")
        return
      })
      src.Window = ws[0]
      ws[0].OnMessage(func(m *astilectron.EventMessage) (v interface{}) {
        opensslHandler := openssl.New()
        var s string
        err := m.Unmarshal(&s)
        if err != nil {
          return "{}"
        }
        info := strings.Split(s, "___::___")
        data, path := src.RequestData{}, info[0]
        if len(info) == 1 {
          data = nil
        } else {
          params := info[1]
          s, err := opensslHandler.DecryptString(src.SecretKey, params)
          if err != nil {
            logger.Errorf("Decrypt Error", err.Error())
            return err.Error()
          }
          if err := json.Unmarshal(s, &data); err != nil {
            logger.Errorf("UnmarshalData Error", err.Error())
            return err.Error()
          }
        }
        if path == "/redis/connection/updatekey" {
          fmt.Println(data)
        }
        return handler.Handle(path, data)
      })
      return nil
    },
    Windows: []*bootstrap.Window{{
      Homepage: url,
      Options: &astilectron.WindowOptions{
        Center:          &center,
        Height:          &height,
        MinHeight:       &height,
        Width:           &width,
        MinWidth:        &width,
        HasShadow:       &HasShadow,
        Fullscreenable:  &FullScreenable,
        Closable:        &Closable,
        AutoHideMenuBar: &skipTaskBar,
      },
    }},
  }

  if err := bootstrap.Run(config); err != nil {
    logger.Errorf("%s", errors.Wrap(err, "running bootstrap failed"))
    os.Exit(0)
  }
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
