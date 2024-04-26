package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"os"
	"os/exec"
	"path"
	"strings"
)

func volumeExtract(volume string) (sourcePath, targetPath string, err error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("volumes parment not enough ")
	}
	sourcePath, targetPath = parts[0], parts[1]
	if sourcePath == "" || targetPath == "" {
		return "", "", fmt.Errorf("invalid volume [%s], path can't be empty", volume)
	}
	return sourcePath, targetPath, nil
}

func mountVolume(mntPath, hostPath, containerPath string) {
	if err := os.Mkdir(hostPath, constant.Perm0777); err != nil {
		log.Infof("mkdir host dir %s error. %v", hostPath, err)
	}
	containerPathInHost := path.Join(mntPath, containerPath)
	if err := os.Mkdir(containerPathInHost, constant.Perm0777); err != nil {
		log.Infof("mkdir host dir %s error. %v", containerPathInHost, err)
	}
	cmd := exec.Command("mount", "-o", "bind", hostPath, containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume failed. %v", err)
	}
}

func unmountVolume(mntPath, containerPath string) {
	containerPathInHost := path.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v", err)
	}
}
