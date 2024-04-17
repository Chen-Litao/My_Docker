package subsystems

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"myself_docker/constant"
	"os"
	"path"
)

type MemorySubSystem struct {
}

// Name 返回cgroup名字
func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.MemoryLimit == "" {
		logrus.Warnf("Memory set is '0' !!!")
		return nil
	}
	subCgroupPath, err := getCgroupPath(cgroupPath, true)
	if err != nil {
		return err
	}
	if res.CpuSet != "" {
		if err = os.WriteFile(path.Join(subCgroupPath, "memory.max"),
			[]byte(res.MemoryLimit),
			constant.Perm0644); err != nil {
			return fmt.Errorf("set cgroup memory fail %v", err)
		}
	}

	return nil
}
