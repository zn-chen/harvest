package dataserver

import (
	"errors"
	"time"
	"log"

	"harvest/protocal/hpfeeds"
)

//runHpfeedsBrokers 运行hpfeedsBroker实例
func runHpfeedsBrokers() chan []byte {
	db := NewDB()
	b := &hpfeeds.Broker{
		Name: "harvestbroker",
		Port: 10000,
		DB:   db,
	}
	b.SetDebugLogger(log.Print)
	b.SetInfoLogger(log.Print)
	b.SetErrorLogger(log.Print)
	go b.ListenAndServe()

	hpfeedChannel := b.LocalSubesriber("pub")
	return hpfeedChannel
}

// AuthDB 认证接口
type AuthDB struct {
	IDs []hpfeeds.Identity
}

//NewDB AuthDB 构造函数
func NewDB() *AuthDB {
	i := hpfeeds.Identity{
		Ident:       "admin",
		Secret:      "admin",
		SubChannels: []string{"sub"},
		PubChannels: []string{"pub"},
	}
	t := &AuthDB{IDs: []hpfeeds.Identity{i}}
	return t
}

// Identify Auth Identify 接口实现
func (a *AuthDB) Identify(ident string) (*hpfeeds.Identity, error) {
	if ident == "admin" {
		return &a.IDs[0], nil
	}
	return nil, errors.New("identifier: Unknown identity")
}


