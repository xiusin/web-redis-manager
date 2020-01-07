package main

import (
	"encoding/json"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/asticode/go-astilectron"
	"github.com/xiusin/redis_manager/src"
)

func main() {
	cacheDir := src.GetCacheDir(src.DEBUG)
	options := astilectron.Options{
		AppName:            "RDM",
		BaseDirectoryPath:  cacheDir,
		DataDirectoryPath:  cacheDir,
		AppIconDefaultPath: cacheDir + "/resources/icon.png",
	}
	a, err := astilectron.New(options)
	if err != nil {
		log.Error(err.Error())
	}
	err = a.Start()
	if err != nil {
		log.Error(err.Error())
	}
	var url string
	if src.DEBUG {
		url = "http://localhost:8899"
	} else {
		url = cacheDir + "/resources/dist/index.html"
	}
	center, HasShadow, Fullscreenable, Closable, MinimizeOnClose := true, true, true, true, true
	height, width := 800, 1280
	w, err := a.NewWindow(url, &astilectron.WindowOptions{
		Center:         &center,
		Height:         &height,
		Width:          &width,
		HasShadow:      &HasShadow,
		Fullscreenable: &Fullscreenable,
		Closable:       &Closable,
		Custom: &astilectron.WindowCustomOptions{
			MinimizeOnClose: &MinimizeOnClose,
		},
	})
	if err != nil {
		log.Error(err.Error())
	}
	w.On(astilectron.EventNameWindowEventMinimize, func(e astilectron.Event) (deleteListener bool) {
		_ = w.Hide()
		return false
	})
  src.Window = w
	err = w.Create()
	if err != nil {
		log.Error(err.Error())
	}
	if src.DEBUG {
		_ = w.OpenDevTools()
	}
	tooltips := "Redis 数据管理工具"
	tr := a.NewTray(&astilectron.TrayOptions{
		Image:   &options.AppIconDefaultPath,
		Tooltip: &tooltips,
	})
	openWin := "打开桌面"
	exitWin := "退出工具"
	m := tr.NewMenu([]*astilectron.MenuItemOptions{
		{
			Label: &openWin,
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				_ = w.Show()
				return false
			},
		},
		{
			Label: &exitWin,
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				_ = a.Quit()
				return false
			},
		},
	})
	_ = tr.Create()
	_ = m.Create()

	tr.On(astilectron.EventNameTrayEventDoubleClicked, func(e astilectron.Event) (deleteListener bool) {
		_ = w.Show()
		return false
	})

	w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var s string
		_ = m.Unmarshal(&s)
		if s == "" {
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
				return err.Error()
			}
		}
		return handler.Handle(info[0], data)
	})
	if runtime.GOOS == "darwin" {
		var d = a.Dock()
		_ = d.Show()
		id, _ := d.Bounce(astilectron.DockBounceTypeCritical)
		_ = d.CancelBounce(id)
		_ = d.SetIcon(options.AppIconDefaultPath)

	}
	a.Wait()
}

var handler *src.Handler

func init() {
	handler = src.NewHandler()

	handler.Add("/redis/connection/test", src.RedisManagerConnectionTest)
	handler.Add("/redis/connection/get-command", src.RedisManagerGetCommandList)
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
}
