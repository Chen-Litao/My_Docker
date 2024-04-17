package subsystems

import (
	"fmt"
	"github.com/sirupsen/logrus"
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
		logrus.Warnf("CpuCfsQuota set is '0' !!!")
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
