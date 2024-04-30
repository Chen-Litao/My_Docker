package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"myself_docker/container"
	"os"
	"path"
	"strconv"
	"syscall"
)

func stopContainer(containerId string) {
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGSTOP); err != nil {
		log.Errorf("Stop container %s error %v", containerId, err)
		return
	}
	containerInfo.Status = container.STOP
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerId, err)
		return
	}
	// 4.重新写回存储容器信息的文件
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	if err := os.WriteFile(configFilePath, newContentBytes, constant.Perm0622); err != nil {
		log.Errorf("Write file %s error:%v", configFilePath, err)
	}

}

func getInfoByContainerId(containerId string) (*container.Info, error) {
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", configFilePath)
	}
	var containerInfo container.Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
