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
	"harvest/collection"
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

	//启动hpfeed服务
	hpSection := conf.Section("HpfeedBroker")
	dataChannel, _ := collection.RunHpfeedsBrokers(
		hpSection.Key("host").MustString("127.0.0.1"),
		hpSection.Key("port").MustInt(10000),
		hpSection.Key("ident").MustString("admin"),
		hpSection.Key("pwd").MustString("admin"),
		hpSection.Key("channel").MustString("default"),
	)
	// go func ()  {
	// 	for {
	// 		<-dataChannel
	// 	}
	// }()

	// 启动rabbitmq客户端
	hpSection = conf.Section("RabbitMQ")
	rabbitmqChannel, err := ConnectionRabbitMQ(
		hpSection.Key("host").MustString("127.0.0.1"), 
		hpSection.Key("port").MustInt(5672), 
		"message", 
		"", 
		hpSection.Key("username").MustString("admin"), 
		hpSection.Key("password").MustString("admin"), 
		"my_vhost",
	)
	if err != nil{
		fmt.Println(err)
		os.Exit(1)
	}
	logger.Info("RabbitMQ connect success")

	go dataPublish(dataChannel, rabbitmqChannel)

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
		logger.Errorln("Watcher faild:", err.Error())
		os.Exit(1)
	}
	w.Start(func() {})
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

// dataPublish 从hpfeed服务器中读取消息，发送到rabbitmq的指定队列中
func dataPublish(dataChannel chan []byte, rabbitmqChannel *Client){
	for data := range dataChannel {
		rabbitmqChannel.Publish(data)
	}
}