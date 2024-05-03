package container

import (
	log "github.com/sirupsen/logrus"
	"myself_docker/constant"
	"myself_docker/utils"
	"os"
	"os/exec"
)

func NewWorkSpace(containerId, volume, imageName string) {
	createLower(containerId, imageName)
	createDirs(containerId)
	mountOverlayFS(containerId)
	if volume != "" {
		//mntpath是为了能够在挂载的时候找到对目录
		mntPath := utils.GetMerged(containerId)
		hostPath, containPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containPath)
	}
}

func DeleteWorkSpace(containerID, volume string) {
	if volume != "" {
		_, continer, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mntPath := utils.GetMerged(containerID)
		unmountVolume(mntPath, continer)
	}
	umountOverlayFS(containerID)
	deleteDirs(containerID)
}

func umountOverlayFS(containerID string) {
	mntPath := utils.GetMerged(containerID)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Infof("umountOverlayFS,cmd:%v", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}
func deleteDirs(containerID string) {
	dirs := []string{
		utils.GetMerged(containerID),
		utils.GetUpper(containerID),
		utils.GetWorker(containerID),
		utils.GetLower(containerID),
		utils.GetRoot(containerID),
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}

func createLower(containerID, imageName string) {
	lowerPath := utils.GetLower(containerID)
	imagePath := utils.GetImage(imageName)

	log.Infof("lowerPath:%s imagePath:%s", lowerPath, imagePath)
	exist, err := utils.PathExists(lowerPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exist %v", lowerPath, err)
	}
	if !exist {
		if err = os.MkdirAll(lowerPath, constant.Perm0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", lowerPath, err)
		}
		if _, err = exec.Command("tar", "-xvf", imagePath, "-C", lowerPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", lowerPath, err)
		}
	}
}
func createDirs(containerID string) {
	dirs := []string{
		utils.GetMerged(containerID),
		utils.GetUpper(containerID),
		utils.GetWorker(containerID),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, constant.Perm0777); err != nil {
			log.Warnf("makedir dir %s error. %v", dir, err)
		}
	}
}

func mountOverlayFS(containerID string) {
	dirs := utils.GetOverlayFSDirs(utils.GetLower(containerID), utils.GetUpper(containerID), utils.GetWorker(containerID))
	mergedPath := utils.GetMerged(containerID)
	//最终指令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	//将三个文件的内容联合成一个文件系统并显示到merged这个目录上
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mergedPath)
	log.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}
