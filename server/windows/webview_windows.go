//go:build windows
// +build windows

package windows

import (
  "fmt"
  "github.com/Tim-Paik/webview2"
  "syscall"
)

var (
  SmCxScreen          = 0
  SmCyScreen          = 1
  dll                 = syscall.MustLoadDLL("user32")
  getSystemMetrics, _ = dll.FindProc("GetSystemMetrics")
)

func InitWebview(port string, isBuild bool) {
  if !isBuild {
    port = "8899"
  }
  if proc, err := dll.FindProc("SetProcessDpiAwarenessContext"); err == nil {
    aware := -4 // 支持HIDPI 其实是根据缩放比处理
    _, _, _ = proc.Call(uintptr(aware))
  }
  w := webview2.New(false)
  if w == nil {
    return
  }
  defer w.Destroy()
  url := fmt.Sprintf("http://localhost:%s/#/", port)
  w.Navigate(url)
  w.SetTitle("Redis Manager")

  width, height := int(float64(GetSystemMetrics(SmCxScreen))*0.7), int(float64(GetSystemMetrics(SmCyScreen))*0.7)
  w.SetSize(width, height, webview2.HintNone)
  //webview2.NewWindow(true, w.Window()) 开启一个新的窗口
  w.Run()
}

func GetSystemMetrics(nIndex int) int {
  index := uintptr(nIndex)
  ret, _, _ := getSystemMetrics.Call(index)
  return int(ret)
}
