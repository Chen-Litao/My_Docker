package cgroups

import (
	"github.com/sirupsen/logrus"
	"myself_docker/cgroups/subsystems"
)

type CgroupManager struct {
	//定义cgroup需要的路径
	Path string
	// 资源限制
	Resource *subsystems.ResourceConfig
}

//创建一个对象
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

//释放cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Apply(c.Path, pid); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}
