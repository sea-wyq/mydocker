package container

import (
	"mydocker/cgroups/subsystems"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func NewWorkSpace(rootURL, mntURL, volume string) {
	log.Infof("createLower")
	createLower(rootURL)
	createDirs(rootURL)
	mountOverlayFS(rootURL, mntURL)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			log.Infof("mountVolume")
			mountVolume(mntURL, volumeURLs)
			log.Infof("volumeURL:%v", volumeURLs)
		} else {
			log.Infof("volume parameter input is not correct.")
		}
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(rootURL string) {
	// 把busybox作为overlayfs中的lower层
	busyboxURL := rootURL + "/busybox"
	busyboxTarURL := rootURL + "/busybox.tar"
	// 检查是否已经存在busybox文件夹
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	// 不存在则创建目录并将busybox.tar解压到busybox文件夹中
	if !exist {
		if err := os.Mkdir(busyboxURL, subsystems.Perm0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("untar dir %s error %v", busyboxURL, err)
		}
	}
}

// createDirs 创建overlayfs需要的的upper、worker目录
func createDirs(rootURL string) {
	upperURL := rootURL + "/upper"
	if err := os.Mkdir(upperURL, subsystems.Perm0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", upperURL, err)
	}
	workURL := rootURL + "/work"
	if err := os.Mkdir(workURL, subsystems.Perm0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", workURL, err)
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(rootURL string, mntURL string) {
	// 创建对应的挂载目录
	if err := os.Mkdir(mntURL, subsystems.Perm0777); err != nil {
		log.Errorf("mountOverlayFS mkdir dir %s error. %v", mntURL, err)
	}
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/merged
	dirs := "lowerdir=" + rootURL + "/busybox" + ",upperdir=" + rootURL + "/upper" + ",workdir=" + rootURL + "/work"
	// 完整命令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	log.Infof("mountOverlayFS cmd:%s", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mountOverlayFS mount err:%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL, mntURL, volume string) {

	// 如果指定了volume则需要先umount volume
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			umountVolume(mntURL, volumeURLs)
		}
	}

	// 然后umount整个容器的挂载点
	umountOverlayFS(mntURL)
	// 最后移除相关文件夹
	deleteDirs(rootURL)
}

func umountOverlayFS(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func deleteDirs(rootURL string) {
	writeURL := rootURL + "/upper"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove dir %s error %v", writeURL, err)
	}
	workURL := rootURL + "/work"
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("remove dir %s error %v", workURL, err)
	}
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
