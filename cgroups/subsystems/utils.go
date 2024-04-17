package subsystems

import (
	"myself_docker/constant"
	"os"
	"path"
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
