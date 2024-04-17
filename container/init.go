package container

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

//容器进程里面启动的第一个程序，目的是将当前的运行目录改为command
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("command:%s", command)
	mountProc()

	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return errors.New("run container get user command error, cmdArray is nil")
	}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

const fdIndex = 3

func readUserCommand() []string {
	//当我们在NewParentProcess中的cmd.ExtraFiles的时候则会传入一个文件描述符从而能够在这里被读到
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	//数据传入采用的是加空格所以输出时通过空格裁开
	return strings.Split(msgStr, " ")
}

func mountProc() {
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	//syscall.MS_NOEXEC  本文件系统中不允许运行其他程序
	//syscall.MS_NOSUID  禁止 setuid 和 setgid 位的更改
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	//为了和宿主机的 /proc 环境隔离，因此在 mydocker init 命令中我们会重新挂载 /proc 文件系统
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
}
