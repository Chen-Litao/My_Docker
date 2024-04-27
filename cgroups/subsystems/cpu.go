package subsystems

import (
	"fmt"
	"myself_docker/constant"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

const (
	PeriodDefault = 100000
	Percent       = 100
)

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuCfsQuota == 0 {
		return nil
	}
	subCgroupPath, err := getCgroupPath(cgroupPath, true)
	if err != nil {
		return err
	}
	if res.CpuCfsQuota != 0 {
		if err = os.WriteFile(path.Join(subCgroupPath, "cpu.max"),
			[]byte(fmt.Sprintf("%s %s", strconv.Itoa(PeriodDefault/Percent*res.CpuCfsQuota), PeriodDefault)),
			constant.Perm0644); err != nil {
			return fmt.Errorf("set cgroup cpu share fail %v", err)

		}
	}
	return nil
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	return applyCgroup(pid, cgroupPath)
}

// Remove 删除cgroupPath对应的cgroup
func (s *CpuSubSystem) Remove(cgroupPath string) error {
	subCgroupPath, err := getCgroupPath(cgroupPath, false)
	if err != nil {
		return err
	}
	return os.Remove(subCgroupPath)
}
