package main

import (
	"fmt"
	"harvest/watcher"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	logger "github.com/sirupsen/logrus"

	"harvest/util"
)

var levelMap = map[string]logger.Level{
	"Trace": logger.TraceLevel,
	"WARN":  logger.WarnLevel,
	"INFO":  logger.InfoLevel,
	"DEBUG": logger.DebugLevel,
	"ERROR": logger.ErrorLevel,
}

func main() {
	var wt *watcher.Watcher
	setupCloseHandler(func() {
		wt.Stop()
	})

	util.ConfigInit("./config.ini")
	conf, _ := util.GetConfig()

	logInt(levelMap[strings.ToUpper(conf.Section("Harvest").Key("level").MustString("INFO"))])
	logger.Info("Harvest process start")

	wt, err := watcher.NewWatcher(
		"ping",
		[]string{"www.baidu.com"},
		10,
		"stderr_log.log",
		"stdout_log.log",
		"",
		"cowrie",
	)
	if err != nil {
		fmt.Println(err)
	}

	go wt.Start(func() {})
	time.Sleep(time.Second * 10)
	wt.Stop()
	time.Sleep(time.Second * 10)
}

func logInt(level logger.Level) {
	logger.SetLevel(level)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp: true,
	})
}

func setupCloseHandler(callback func()) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		callback()
		time.Sleep(time.Second * 1)
		os.Exit(0)
	}()
}
