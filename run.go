package main

import (
	log "github.com/sirupsen/logrus"
	"myself_docker/cgroups"
	"myself_docker/cgroups/subsystems"
	"myself_docker/container"
	"myself_docker/network"
	"os"
	"strconv"
	"strings"
)

func Run(tty bool, envSlice, comArray []string, resConf *subsystems.ResourceConfig, volume, containerName, imageName string, net string, portMapping []string) {
	containerId := container.GenerateContainerID() // 生成 10 位容器 id

	parent, writePipe := container.NewParentProcess(tty, volume, containerId, imageName, envSlice) //采用管道进行消息传递
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil { //启动子进程创建
		log.Error(err)
	}
	//记录容器信息
	err := container.RecordContainerInfo(parent.Process.Pid, comArray, containerName, containerId, imageName, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(resConf)
	_ = cgroupManager.Apply(parent.Process.Pid)

	if net != "" {
		// config container network
		containerInfo := &container.Info{
			Id:          containerId,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		if _, err = network.Connect(net, containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}

	sendInitCommand(comArray, writePipe)
	if tty {
		_ = parent.Wait()
		container.DeleteWorkSpace(containerId, volume)
		container.DeleteContainerInfo(containerId)
	}

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
