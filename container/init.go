package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

//容器进程里面启动的第一个程序，目的是将当前的运行目录改为
func RunContainerInitProcess(cmd string) error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	//设置模式，
	//syscall.MS_NOEXEC  本文件系统中不允许运行其他程序
	//syscall.MS_NOSUID  禁止 setuid 和 setgid 位的更改
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
