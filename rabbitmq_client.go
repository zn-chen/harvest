package main

import (
	"fmt"

	"github.com/streadway/amqp"

	logger "github.com/sirupsen/logrus"
)

var rabbitmqConn *amqp.Connection

//Client rabbitmq消息发布客户端
type Client struct {
	Host     string
	Port     int
	exchange string
	routeKey string
	username string
	password string
	ch       *amqp.Channel
	close    bool
}

// ConnectionRabbitMQ 连接rabbitmq
func ConnectionRabbitMQ(host string, port int, exchange string, routeKey string, username string, password string, vhost string) (*Client, error) {
	if rabbitmqConn == nil{
		url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, vhost)
		var err error
		rabbitmqConn, err = amqp.Dial(url)
		if err != nil {
			logger.Errorln("Rabbimq connection faild")
			return nil, err
		}
	}

	ch, err := rabbitmqConn.Channel()
	if err != nil {
		logger.Errorln("Rabbitmq get channel faild")
		return nil, err
	}

	return &Client{
		Host:     host,
		Port:     port,
		exchange: exchange,
		routeKey: routeKey,
		username: username,
		password: password,
		ch:       ch,
		close:    false,
	}, nil
}

//Publish 发布消息
func (p *Client) Publish(body []byte) error {
	err := p.ch.Publish(
		p.exchange,
		p.routeKey,
		false, // mandatory
		false, // mandatory
		amqp.Publishing{
			ContentType: "text/plain",
			Body: body,
		},
	)
	if err != nil {
		logger.Error("Publisher faild")
		return err
	}
	return nil
}
