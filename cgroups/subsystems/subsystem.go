package subsystems

type ResourceConfig struct {
	MemoryLimit string
	CpuCfsQuota int
	CpuSet      string
}

//定义对子系统的接口
type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
