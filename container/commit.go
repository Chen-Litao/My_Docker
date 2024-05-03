package container

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"myself_docker/utils"
	"os/exec"
)

func CommitContainer(containerID, imageName string) error {
	mntPath := utils.GetMerged(containerID)
	imageTar := utils.GetImage(imageName)
	exists, err := utils.PathExists(imageTar)
	if err != nil {
		return errors.WithMessagef(err, "check is image [%s/%s] exist failed", imageName, imageTar)
	}
	if exists {
		return errors.New("Image Already Exists")
	}
	fmt.Println("commitContainer imageTar:", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntPath, err)
		return err
	}
	return nil
}
