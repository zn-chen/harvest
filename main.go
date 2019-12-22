package main

import (
    "harvest/watcher"
	"time"
	"fmt"
)

func main()  {
    wt, err := watcher.NewWatcher(
        "ping",
        []string{"www.baidu.com"},
        10,
        "stderr_log.log",
        "stdout_log.log",
        "",
        "",
    )
    if err != nil {
        fmt.Println(err)
    }
    
    go wt.Start(func() {})
    time.Sleep(time.Second * 10)
    wt.Stop()
    time.Sleep(time.Second * 10)
}