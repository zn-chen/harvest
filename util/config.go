package util

import (
	"errors"
	"gopkg.in/ini.v1"
)

//GlobalConfiguration 全局配置
type GlobalConfiguration = ini.File

var instance *GlobalConfiguration

//GetConfig 获取一个全局配置
func GetConfig() (*GlobalConfiguration, error) {
	if instance == nil {
		return nil, errors.New("配置未初始化")
	}
	return instance, nil
}

//ConfigInit 配置文件初始化
func ConfigInit(path string) (*GlobalConfiguration, error) {
	file, err := ini.Load(path)
	instance = file
	return file, err
}