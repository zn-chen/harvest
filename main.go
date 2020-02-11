package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"harvest/watcher"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	logger "github.com/sirupsen/logrus"

	"harvest/collection"
	"harvest/util"
)

var levelMap = map[string]logger.Level{
	"Trace": logger.TraceLevel,
	"WARN":  logger.WarnLevel,
	"INFO":  logger.InfoLevel,
	"DEBUG": logger.DebugLevel,
	"ERROR": logger.ErrorLevel,
}

//Ident 节点标识
var Ident string
//NodeName 节点名称
var NodeName string
// ServerHost 中心地址
var ServerHost string
// ServerPort 中心服务端口
var ServerPort int
// Describe 节点描述
var Describe string

func main() {
	//注册退出signal
	setupCloseHandler(func() {})

	//初始化配置文件
	util.ConfigInit("./config.ini")
	conf, _ := util.GetConfig()

	//初始化日志
	logInit(levelMap[strings.ToUpper(conf.Section("Harvest").Key("level").MustString("INFO"))])

	// 初始化公共变量
	Ident = conf.Section("Harvest").Key("ident").MustString("0000000000000000")
	NodeName = conf.Section("Harvest").Key("name").MustString("uname")
	ServerHost = conf.Section("Server").Key("host").MustString("127.0.0.1")
	ServerPort = conf.Section("Server").Key("port").MustInt(80)
	Describe = conf.Section("Server").Key("describe").MustString("")

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
		"collector",
		"",
		hpSection.Key("username").MustString("admin"),
		hpSection.Key("password").MustString("admin"),
		"message",
	)
	if err != nil {
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

	//向中心注册节点
	register(
		Ident,
		NodeName,
		ServerHost,
		ServerPort,
		Describe,
	)
	// 开始发送心跳
	go beat()

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
func dataPublish(dataChannel chan []byte, rabbitmqChannel *Client) {
	for data := range dataChannel {
		data, err := normalize(data)
		if err != nil {
			logger.Errorf("Data normalize error %s", err)
		}
		rabbitmqChannel.Publish(data)
	}
}

// RegisterPost 注册用的post
type RegisterPost struct {
	Ident    string `json:"ident"`
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	Describe string `json:"describe"`
}

// 向中心注册节点
func register(ident string, name string, addr string, port int, describe string) {
	data := RegisterPost{
		ident,
		name,
		addr,
		port,
		describe,
	}
	url := fmt.Sprintf("http://%s:%d/api/v1.0/nodes", ServerHost, ServerPort)
	byteData, _ := json.Marshal(data)
	reader := bytes.NewReader(byteData)

	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		logger.Errorf("Regist faild %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Errorf("Regist Faild code %d", resp.StatusCode)
	} else {
		logger.Infof("Regist success")
	}
}

// 心跳任务
func beat() {
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/v1.0/beat/%s", ServerHost, ServerPort, Ident))
		if err != nil {
			logger.Errorf("Beat request Faild code %s", err)
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			logger.Errorf("Beat request Faild code %d", resp.StatusCode)
		} else {
			logger.Debugf("Beat request success")
		}

		time.Sleep(30 * time.Second)
	}
}
