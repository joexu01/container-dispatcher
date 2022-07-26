package main

import (
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	_ = lib.InitModule("./conf/dev/", []string{"base", "redis", "mysql"})
	defer lib.Destroy()
	router.HttpsServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	router.HttpsServerStop()
}
