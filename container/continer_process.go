package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"myself_docker/utils"
	"os"
	"os/exec"
	"syscall"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/lib/mydocker/containers/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
	LogFile       = "%s-json.log"
)

type Info struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
	ImageName   string `json:"imageName"`  // 容器采用的镜像名称
	Volume      string `json:"volume"`     // 容器挂载的 volume
}

//创建一个新的实现隔离的容器进程
func NewParentProcess(tty bool, volume, containerId, imageName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirPath := fmt.Sprintf(InfoLocFormat, containerId)
		if err := os.MkdirAll(dirPath, constant.Perm0622); err != nil {
			log.Errorf("NewParentProcess mkdir %s error %v", dirPath, err)
			return nil, nil
		}
		stdLogFilePath := dirPath + GetLogfile(containerId)
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile

	}
	cmd.ExtraFiles = []*os.File{readPipe}

	NewWorkSpace(containerId, volume, imageName)
	//表示最后交由用户看到的是merged这个目录
	cmd.Dir = utils.GetMerged(containerId)
	return cmd, writePipe
}
