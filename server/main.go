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

const DEBUG = true

var once sync.Once

var cacheDir string

var handler = src.NewHandler()

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
    "/redis/connection/pubsub":      src.RedisPubSub,
    "/redis/connection/info":        src.RedisManagerGetInfo,
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
