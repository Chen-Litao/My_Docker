package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func CommitContainer(imageName string) error {
	mntPath := "/root/merged"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Println("commitContainer imageTar:", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntPath, err)
		return err
	}
	return nil
}
