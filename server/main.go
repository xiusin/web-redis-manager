package main

import (
  "encoding/json"
  "fmt"
  "github.com/asticode/go-astilectron"
  bootstrap "github.com/asticode/go-astilectron-bootstrap"
  "github.com/asticode/go-astilog"
  "github.com/pkg/errors"
  "github.com/xiusin/redis_manager/server/src"
  "log"
  "os"
  "path/filepath"
  "strings"
  "sync"
)

const DEBUG  =  true

var once sync.Once

var cacheDir string

var handler = src.NewHandler()

func init() {
  cacheDir = GetCacheDir()

  astilog.SetLogger(astilog.New(astilog.Configuration{
    AppName:  "RedisManager",
    Filename: cacheDir + `/rdm-log.log`,
    Verbose:  DEBUG,
  }))
  astilog.FlagConfig()

  handler.Add("/redis/connection/test", src.RedisManagerConnectionTest)
  handler.Add("/redis/connection/save", src.RedisManagerConfigSave)
  handler.Add("/redis/connection/list", src.RedisManagerConnectionList)
  handler.Add("/redis/connection/server", src.RedisManagerConnectionServer)
  handler.Add("/redis/connection/removekey", src.RedisManagerRemoveKey)
  handler.Add("/redis/connection/removerow", src.RedisManagerRemoveRow)
  handler.Add("/redis/connection/updatekey", src.RedisManagerUpdateKey)
  handler.Add("/redis/connection/addkey", src.RedisManagerAddKey)
  handler.Add("/redis/connection/flushDB", src.RedisManagerFlushDB)
  handler.Add("/redis/connection/remove", src.RedisManagerRemoveConnection)
  handler.Add("/redis/connection/command", src.RedisManagerCommand)
  handler.Add("/redis/connection/pubsub", src.RedisPubSub)
  handler.Add("/redis/connection/info", src.RedisManagerGetInfo)
  handler.Add("/redis/connection/get-command", src.RedisManagerGetCommandList)
}


func main() {
  options := astilectron.Options{
    AppName:            "RedisManager",
    SingleInstance:     true,
    BaseDirectoryPath:  cacheDir,
    AppIconDefaultPath: fmt.Sprintf("%s/resources/icon.png", cacheDir),
    DataDirectoryPath:  cacheDir,
    //VersionElectron:    "6.1.2",
  }

  var url string
  if DEBUG {
    url = "http://localhost:8899"
  } else {
    url = "index.html"
  }
  //else {
  //  url = cacheDir + "/resources/dist/index.html"
  //}
  center, HasShadow, Fullscreenable, Closable, skipTaskBar := true, true, true, true, true
  height, width := 800, 1280

  config := bootstrap.Options{
    //Asset:              Asset,
    //AssetDir:           AssetDir,
    AstilectronOptions: options,
    Debug:              DEBUG,
    Logger:             astilog.GetLogger(),
    //RestoreAssets:      RestoreAssets,
    OnWait: func(a *astilectron.Astilectron, ws []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
      a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
        log.Println("App has crashed")
        return
      })
      src.Window = ws[0]
      ws[0].OnMessage(func(m *astilectron.EventMessage) (v interface{}) {
        var s string
        err := m.Unmarshal(&s)
        if err != nil {
          return "{}"
        }

        info := strings.Split(s, "___::___")
        data := make(map[string]interface{})
        if len(info) == 1 {
          data = nil
        } else {
          err := json.Unmarshal([]byte(info[1]), &data)
          if err != nil {
            astilog.Errorf("UnmarshalData Error", err.Error())
            return err.Error()
          }
        }
        return handler.Handle(info[0], data)
      })

      return nil
    },
    Windows: []*bootstrap.Window{{
      Homepage: url,
      Options: &astilectron.WindowOptions{
        Center:          &center,
        Height:          &height,
        Width:           &width,
        HasShadow:       &HasShadow,
        Fullscreenable:  &Fullscreenable,
        Closable:        &Closable,
        AutoHideMenuBar: &skipTaskBar,
      },
    }},
  }

  if err := bootstrap.Run(config); err != nil {
    astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
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
