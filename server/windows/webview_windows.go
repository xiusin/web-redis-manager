//go:build windows
// +build windows

package windows

import (
	"fmt"
	"github.com/jchv/go-webview2"
	"syscall"
)

var (
	SM_CXSCREEN         = 0
	SM_CYSCREEN         = 1
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
	w := webview2.New(true)
	width, height := int(float64(GetSystemMetrics(SM_CXSCREEN))*0.65), int(float64(GetSystemMetrics(SM_CYSCREEN))*0.65)
	w.SetSize(width, height, webview2.HintFixed)
	if w == nil {
		return
	}
	defer w.Destroy()
	url := fmt.Sprintf("http://localhost:%s/#/", port)
	w.SetTitle("RedisManager")
	w.Navigate(url)
	w.Run()
}

func GetSystemMetrics(nIndex int) int {
	index := uintptr(nIndex)
	ret, _, _ := getSystemMetrics.Call(index)
	return int(ret)
}
