package container

import (
	"fmt"
	"mydocker/cgroups/subsystems"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

/*
*
Init 挂载点
*/
func setUpMount() {
	pwd, err := os.Getwd() // 在进程中设置的路径 mntURL
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	if err = pivotRoot(pwd); err != nil {
		log.Errorf("pivotRoot error %v", err)
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("mount proc error %v", err)
	}
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		log.Errorf("mount tmpfs error %v", err)
	}
}

func pivotRoot(root string) error {
	/**
	   PivotRoot调用有限制，newRoot和oldRoot不能在同一个文件系统下，因此，为了使当前root的老root和新root不在同一个文件系统下，
	这里把root重新mount了一次。
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}
	// 创建 rootfs/.pivot_root 目录用于存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, subsystems.Perm0777); err != nil {
		return err
	}
	// 执行pivot_root调用,将系统rootfs切换到新的rootfs,
	// PivotRoot调用会把 old_root挂载到pivotDir,也就是rootfs/.pivot_root,挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	// 最后再把old_root umount了，即 umount rootfs/.pivot_root
	// 由于当前已经是在 rootfs 下了，就不能再用上面的rootfs/.pivot_root这个路径了,现在直接用/.pivot_root这个路径即可
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
