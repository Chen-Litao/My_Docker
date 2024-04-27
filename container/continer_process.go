package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/lib/mydocker/containers/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
)

type Info struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
}

//创建一个新的实现隔离的容器进程
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	//cmd := exec.Command("/proc/self/exe", "init")
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	Cloneflags: syscall.CLONE_NEWUTS |
	//		syscall.CLONE_NEWPID |
	//		syscall.CLONE_NEWNS |
	//		syscall.CLONE_NEWNET |
	//		syscall.CLONE_NEWIPC,
	//}
	//if tty {
	//	cmd.Stdin = os.Stdin
	//	cmd.Stdout = os.Stdout
	//	cmd.Stderr = os.Stderr
	//}
	cmd.ExtraFiles = []*os.File{readPipe}
	rootPath := "/root"
	NewWorkSpace(rootPath, volume)
	//表示最后交由用户看到的是merged这个目录
	cmd.Dir = path.Join(rootPath, "merged")
	return cmd, writePipe
}

func NewWorkSpace(rootPath, volume string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath)
	if volume != "" {
		//mntpath是为了能够在挂载的时候找到对目录
		mntPath := path.Join(rootPath, "merged")
		hostPath, containPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containPath)
	}
}

func DeleteWorkSpace(rootPath, volume string) {
	mntPath := path.Join(rootPath, "merged")
	if volume != "" {
		_, continer, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		unmountVolume(mntPath, continer)
	}
	umountOverlayFS(mntPath)
	deleteDirs(rootPath)
}

func umountOverlayFS(mntPath string) {
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Infof("umountOverlayFS,cmd:%v", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}
func deleteDirs(rootPath string) {
	dirs := []string{
		path.Join(rootPath, "merged"),
		path.Join(rootPath, "upper"),
		path.Join(rootPath, "work"),
	}

	for _, dir := range dirs {
		if err := os.Remove(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}

func createLower(rootPath string) {
	busyboxPath := path.Join(rootPath, "busybox")
	busyboxTarPath := path.Join(rootPath, "busybox.tar")
	log.Infof("busybox:%s busybox.tar:%s", busyboxPath, busyboxTarPath)
	exist, err := PathExists(busyboxPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exist %v", busyboxPath, err)
	}
	if !exist {
		if err = os.Mkdir(busyboxPath, constant.Perm0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxPath, err)
		}
		if _, err = exec.Command("tar", "-xvf", busyboxTarPath, "-C", busyboxPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxPath, err)
		}
	}

}
func createDirs(rootPath string) {
	dirs := []string{
		path.Join(rootPath, "merged"),
		path.Join(rootPath, "upper"),
		path.Join(rootPath, "work"),
	}
	for _, dir := range dirs {
		if err := os.Mkdir(dir, constant.Perm0777); err != nil {
			log.Warnf("makedir dir %s error. %v", dir, err)
		}
	}
}

func mountOverlayFS(rootPath string) {
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", path.Join(rootPath, "busybox"),
		path.Join(rootPath, "upper"), path.Join(rootPath, "work"))
	//最终指令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	//将三个文件的内容联合成一个文件系统并显示到merged这个目录上
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, path.Join(rootPath, "merged"))
	log.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func PathExists(rootPath string) (bool, error) {
	_, err := os.Stat(rootPath)
	if err == nil {
		return true, err
	}
	if os.IsNotExist(err) {
		return false, err
	}
	return false, err
}
