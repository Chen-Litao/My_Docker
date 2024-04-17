package container

import (
	"os"
	"os/exec"
	"syscall"
)

//创建一个新的实现隔离的容器进程
func NewParentProcess(tty bool, command string) *exec.Cmd {

	args := []string{"init", command}
	//路径：/proc/self/exe表示自我调用在使用时表示自己调用自己正在执行的程序
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
