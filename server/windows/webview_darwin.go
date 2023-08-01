package windows

import (
	"github.com/webview/webview"
)

func InitWebview(url string) {
	w := webview.New(false)
	defer w.Destroy()
	w.SetSize(1200, 800, webview.HintNone)
	w.SetTitle("Redis Manager")
	w.Navigate(url)
	w.Run()
}
