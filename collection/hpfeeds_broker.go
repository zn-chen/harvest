package collection

import (
	"errors"

	// logger "github.com/sirupsen/logrus"

	"harvest/protocal/hpfeeds"
	// "harvest/util"
)

// hpfeed消息采集器,在本地运行一个单用户的hpfeed broker服务，用于接收支持hpfeeds消息格式的蜜罐

//RunHpfeedsBrokers 运行hpfeedsBroker实例
func RunHpfeedsBrokers(host string, port int, ident string, pwd string, channel string) chan []byte {
	db := NewDB(ident, pwd, channel)
	b := hpfeeds.NewBroker(
		host, 
		"harvest", 
		port, 
		db,
	)
	
	hpfeedChannel := b.LocalSubesriber(channel)
	
	go b.ListenAndServe()
	return hpfeedChannel
}

// AuthDB 认证接口
type AuthDB struct {
	ident string
	IDs []hpfeeds.Identity
}

//NewDB AuthDB 构造函数
func NewDB(ident, pwd, channel string) *AuthDB {

	i := hpfeeds.Identity{
		Ident:       ident,
		Secret:      pwd,
		SubChannels: []string{},
		PubChannels: []string{channel},
	}
	t := &AuthDB{ident:ident, IDs: []hpfeeds.Identity{i}}
	return t
}

// Identify Auth Identify 接口实现
func (a *AuthDB) Identify(ident string) (*hpfeeds.Identity, error) {
	if ident == a.ident {
		return &a.IDs[0], nil
	}
	return nil, errors.New("identifier: Unknown identity")
}


