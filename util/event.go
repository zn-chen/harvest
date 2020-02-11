
package util

//Event python asyncio event 的go实现
type Event struct {
	isSet        bool
	eventChannel chan bool
}

//NewEvent 创建并返回一个事件
func NewEvent() *Event {
	return &Event{
		isSet:        false,
		eventChannel: make(chan bool),
	}
}

//IsSet 判断该事件是否被设置
func (e *Event) IsSet() bool {
	return e.isSet
}

//Set 设置该事件
func (e *Event) Set() {
	if e.isSet {
		return
	}
	e.isSet = true
	close(e.eventChannel)
}

//Clear 重置该事件
func (e *Event) Clear() {
	if !e.isSet {
		return
	}
	e.isSet = false
	e.eventChannel = make(chan bool)
}

//Wait 阻塞直到该事件被设置
func (e *Event) Wait() {
	<-e.eventChannel
}
