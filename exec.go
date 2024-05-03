package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"myself_docker/container"
	_ "myself_docker/nsenter"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerId string, comArray []string) {
	pid, err := getPidByContainerId(containerId)
	if err != nil {
		log.Errorf("Exec container getContainerPidByName %s error %v", containerId, err)
		return
	}
	containerEnvs := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnvs...)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 把命令拼接成字符串，便于传递
	cmdStr := strings.Join(comArray, " ")
	log.Infof("container pid：%s command：%s", pid, cmdStr)
	_ = os.Setenv(EnvExecPid, pid)
	_ = os.Setenv(EnvExecCmd, cmdStr)

	if err = cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerId, err)
	}
}

func getPidByContainerId(containerId string) (string, error) {
	//先拼接出获取数据的对象
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	//读取内容并解析
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("Read file %s error %v", path, err)
		return nil
	}
	// env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
