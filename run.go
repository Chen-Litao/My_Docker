package main

import (
	log "github.com/sirupsen/logrus"
	"myself_docker/cgroups"
	"myself_docker/cgroups/subsystems"
	"myself_docker/container"
	"os"
	"strings"
)

func Run(tty bool, comArray []string, resConf *subsystems.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	sendInitCommand(comArray, writePipe)
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(resConf)
	_ = cgroupManager.Apply(parent.Process.Pid)
	_ = parent.Wait()
	container.DeleteWorkSpace("/root/", volume)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
