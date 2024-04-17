package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

//容器进程里面启动的第一个程序，目的是将当前的运行目录改为command
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("command:%s", command)

	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示声明你要这个新的mount namespace独立。
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	//设置模式，
	//syscall.MS_NOEXEC  本文件系统中不允许运行其他程序
	//syscall.MS_NOSUID  禁止 setuid 和 setgid 位的更改
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	//为了和宿主机的 /proc 环境隔离，因此在 mydocker init 命令中我们会重新挂载 /proc 文件系统
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
