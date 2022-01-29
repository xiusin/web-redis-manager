//go:build !windows

package windows

import (
	"os"
	"os/signal"
)

func InitWebview(port int, isBuild bool) {
	signalCh := make(os.Signal)
	signal.Notify(signalCh, os.Kill, os.Interrupt)
	<-signalCh
}
