package container

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func RemoveContainer(containerName string) {
	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	// 限制只能删除STOP状态的容器
	if containerInfo.Status != STOP {
		log.Errorf("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(InfoLocFormat, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove file %s error %v", dirURL, err)
		return
	}
}
