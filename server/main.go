package main

import (
  "encoding/json"
  "github.com/asticode/go-astilectron"
  bootstrap "github.com/asticode/go-astilectron-bootstrap"
  "github.com/asticode/go-astilog"
  "github.com/pkg/errors"
  "github.com/xiusin/redis_manager/server/src"
  "strings"
)

func main() {
  cacheDir := src.GetCacheDir(src.DEBUG)
  options := astilectron.Options{
    AppName:            "RedisManager",
    SingleInstance:     true,
    BaseDirectoryPath:  cacheDir,
    AppIconDefaultPath: cacheDir + "/resources/icon.png",
    DataDirectoryPath:  cacheDir,
  }
  // 启动内部端口监听服务
  var url string
  if src.DEBUG {
    url = "http://localhost:8899"
  } else {
    url = "index.html"
  }
  //else {
  //  url = cacheDir + "/resources/dist/index.html"
  //}
  astilog.Infof("url: %s", url)
  center, HasShadow, Fullscreenable, Closable, skipTaskBar := true, true, true, true, true
  height, width := 800, 1280

  config := bootstrap.Options{
    //Asset:              Asset,
    //AssetDir:           AssetDir,
    AstilectronOptions: options,
    Debug:              true,
    Logger:             astilog.GetLogger(),
    //RestoreAssets:      RestoreAssets,
    OnWait: func(_ *astilectron.Astilectron, ws []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
      src.Window = ws[0]
      ws[0].OnMessage(func(m *astilectron.EventMessage) (v interface{}) {
        var s string
        err := m.Unmarshal(&s)
        if err != nil {
          return "{}"
        }

        //拆分路由以及数据内容
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
        Custom: &astilectron.WindowCustomOptions{

        },
      },
    }},
  }

  if err := bootstrap.Run(config); err != nil {
    astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
  }
}

var handler *src.Handler

func init() {
  handler = src.NewHandler()

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
