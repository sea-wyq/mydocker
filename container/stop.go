package container

import (
	"encoding/json"
	"fmt"
	"mydocker/cgroups/subsystems"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func StopContainer(containerName string) {
	// 1. 根据容器名称获取对应 PID
	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("get container %s info error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("conver pid from string to int error %v", err)
		return
	}
	// 2.发送SIGTERM信号
	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("stop container %s error %v", containerName, err)
		return
	}
	// 3.修改容器信息，将容器置为STOP状态，并清空PID
	containerInfo.Status = STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("json marshal %s error %v", containerName, err)
		return
	}
	// 4.重新写回存储容器信息的文件
	dirURL := fmt.Sprintf(InfoLocFormat, containerName)
	configFilePath := dirURL + ConfigName
	if err := os.WriteFile(configFilePath, newContentBytes, subsystems.Perm0622); err != nil {
		log.Errorf("write file %s error:%v", configFilePath, err)
	}
}

func GetContainerInfoByName(containerName string) (*Info, error) {
	dirURL := fmt.Sprintf(InfoLocFormat, containerName)
	configFilePath := dirURL + ConfigName
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", configFilePath)
	}
	var containerInfo Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
