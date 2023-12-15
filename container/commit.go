package container

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func CommitContainer(imageName string) {
	imageTar := RootURL + "/images/" + imageName + ".tar"
	fmt.Println("commitContainer imageTar:", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", MntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", MntURL, err)
	}
}
