package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	logger "github.com/sirupsen/logrus"

	"harvest/util"
)

//Watcher 进程托管
type Watcher struct {
	breakEvent    *util.Event
	tryRestartNum int
	cmd           string
	directory     string
	user          string
	group         string
	Process       *exec.Cmd
	stdoutFile    string
	stderrFile    string
}

//NewWatcher 创建一个进程托管
func NewWatcher(cmd string, arg []string, tryNum int, stderrFile, stdoutFile, directory, runInUser string) (*Watcher, error) {
	process := exec.Command(cmd, arg...)

	//设定工作目录, 如为空则为当前运行目录
	if directory != "" {
		process.Dir = directory
	}

	//输出重定向
	if stderrFile != "" {
		stderrIO, _ := os.OpenFile(stderrFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		process.Stderr = stderrIO
	}
	if stdoutFile != "" {
		// stdoutIO, _ := os.Create(stdoutFile)
		stdoutIO, _ := os.OpenFile(stdoutFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		process.Stdout = stdoutIO
	}

	//切换用户和组
	if runInUser != "" {
		usr, err := user.Lookup(runInUser)
		if err != nil {
			return nil, fmt.Errorf("User:%s does not exist", runInUser)
		}
		uid, _ := strconv.Atoi(usr.Uid)
		gid, _ := strconv.Atoi(usr.Gid)

		process.SysProcAttr = &syscall.SysProcAttr{}
		process.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	return &Watcher{
		breakEvent:    util.NewEvent(),
		tryRestartNum: tryNum,
		cmd:           cmd,
		directory:     directory,
		user:          runInUser,
		Process:       process,
	}, nil
}

//Start 运行托管程序
func (w *Watcher) Start(callback func()) error {
	logger.Info("Watcher start")
	w.breakEvent.Clear()
	var err error
	for !w.breakEvent.IsSet() {
		err = w.tryRestart()
		if err != nil {
			break
		}
		err = w.Process.Wait()
		if err == nil {
			break
		}
	}
	w.breakEvent.Set()
	logger.Info("Child process exit")
	callback()
	return err
}

func (w *Watcher) tryRestart() error {
	var i int
	var err error
	for i = 0; i < w.tryRestartNum && !w.breakEvent.IsSet(); i++ {
		err = w.Process.Start()
		if err == nil {
			break
		}
	}

	if i < w.tryRestartNum {
		return nil
	}
	return fmt.Errorf("Can not run this com, after %d th try. %s", i, err)
}

//Stop 停止托管
func (w *Watcher) Stop() {
	w.breakEvent.Set()
	w.Process.Process.Signal(syscall.SIGINT)
	logger.Debug("Send SIGINT to child process")
	w.breakEvent.Wait()
}
