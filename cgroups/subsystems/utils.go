package subsystems

import (
	"fmt"
	"myself_docker/constant"
	"os"
	"path"
	"strconv"
)

const UnifiedMountpoint = "/sys/fs/cgroup"

func getCgroupPath(cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := UnifiedMountpoint
	absPath := path.Join(cgroupRoot, cgroupPath)
	if !autoCreate {
		return absPath, nil
	}
	_, err := os.Stat(absPath)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(absPath, constant.Perm0755)
		return absPath, err
	}
	return absPath, err
}

func applyCgroup(pid int, cgroupPath string) error {
	subCgroupPath, err := getCgroupPath(cgroupPath, true)
	if err != nil {
		return fmt.Errorf("get cgroup fail %v", err)
	}
	if err = os.WriteFile(path.Join(subCgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)),
		constant.Perm0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil

}
