package container

import (
	"mydocker/cgroups/subsystems"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func mountVolume(mntURL string, volumeURLs []string) {
	// 第0个元素为宿主机目录
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, subsystems.Perm0777); err != nil {
		log.Infof("mkdir parent dir %s error. %v", parentUrl, err)
	}
	// 第1个元素为容器目录
	containerUrl := volumeURLs[1]
	// 拼接并创建对应的容器目录
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.Mkdir(containerVolumeURL, subsystems.Perm0777); err != nil {
		log.Infof("mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// 通过bind mount 将宿主机目录挂载到容器目录
	// mount -o bind /hostURL /containerURL
	cmd := exec.Command("mount", "-o", "bind", parentUrl, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume failed. %v", err)
	}
}

func umountVolume(mntURL string, volumeURLs []string) {
	containerUrl := mntURL + "/" + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume failed. %v", err)
	}
}

// volumeUrlExtract 通过冒号分割解析volume目录，比如 -v /tmp:/tmp
func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}
