package subsystems

import (
	"fmt"
	"myself_docker/constant"
	"os"
	"path"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}
	subCgroupPath, err := getCgroupPath(cgroupPath, true)
	if err != nil {
		return err
	}
	if res.CpuSet != "" {
		if err = os.WriteFile(path.Join(subCgroupPath, "cpuset.cpus"),
			[]byte(res.CpuSet),
			constant.Perm0644); err != nil {
			return fmt.Errorf("set cgroup cpuset fail %v", err)
		}
	}

	return nil
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	return applyCgroup(pid, cgroupPath)
}

// Remove 删除cgroupPath对应的cgroup
func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	subCgroupPath, err := getCgroupPath(cgroupPath, false)
	if err != nil {
		return err
	}
	return os.Remove(subCgroupPath)
}
