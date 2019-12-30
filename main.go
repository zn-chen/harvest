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
	//注册退出signal
	setupCloseHandler(func() {})

	//初始化配置文件
	util.ConfigInit("./config.ini")
	conf, _ := util.GetConfig()

	//初始化日志
	logInit(levelMap[strings.ToUpper(conf.Section("Harvest").Key("level").MustString("INFO"))])

	//启动进程托管
	commond := strings.Split(conf.Section("Wather").Key("cmd").String(), " ")
	tryNum, _ := conf.Section("Wather").Key("try_num").Int()
	w, err := watcher.NewWatcher(
		commond[0],
		commond[1:],
		tryNum,
		conf.Section("Wather").Key("stderr").String(),
		conf.Section("Wather").Key("stdout").String(),
		conf.Section("Wather").Key("directory").String(),
		conf.Section("Wather").Key("user").String(),
	)
	defer w.Stop()
	if err != nil {
		logger.Errorln("Watcher faild: %s", err)
	}

	go w.Start(func() {})
	logger.Infoln("Watcher running")

	logger.Info("Harvest process start")
}

func logInit(level logger.Level) {
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