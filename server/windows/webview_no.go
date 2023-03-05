//go:build !windows

package windows

import (
	"os"
	"os/signal"
	"syscall"
)

func InitWebview(port int, isBuild bool) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGALRM)
	<-signalCh
}
