package windows

import (
	"syscall"

	"github.com/Tim-Paik/webview2"
)

var (
	SmCxScreen          = 0
	SmCyScreen          = 1
	dll                 = syscall.MustLoadDLL("user32")
	getSystemMetrics, _ = dll.FindProc("GetSystemMetrics")
)

func InitWebview(url string) {
	if proc, err := dll.FindProc("SetProcessDpiAwarenessContext"); err == nil {
		aware := -4 // 支持HIDPI 其实是根据缩放比处理
		_, _, _ = proc.Call(uintptr(aware))
	}
	w := webview2.New(false)
	defer w.Destroy()
	w.Navigate(url)
	w.SetTitle("Redis Manager")

	width, height := int(float64(GetSystemMetrics(SmCxScreen))*0.7), int(float64(GetSystemMetrics(SmCyScreen))*0.7)
	w.SetSize(width, height, webview2.HintNone)
	w.Run()
}

func GetSystemMetrics(nIndex int) int {
	index := uintptr(nIndex)
	ret, _, _ := getSystemMetrics.Call(index)
	return int(ret)
}
