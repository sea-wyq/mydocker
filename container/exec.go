package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	// 需要导入nsenter包，以触发C代码
	_ "mydocker/nsenter"

	log "github.com/sirupsen/logrus"
)

// nsenter里的C代码里已经出现mydocker_pid和mydocker_cmd这两个Key,主要是为了控制是否执行C代码里面的setns.
const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerName string, comArray []string) {
	// 根据传进来的容器名获取对应的PID
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 把命令拼接成字符串，便于传递
	cmdStr := strings.Join(comArray, " ")
	log.Infof("container pid:%s command:%s", pid, cmdStr)
	_ = os.Setenv(EnvExecPid, pid)
	_ = os.Setenv(EnvExecCmd, cmdStr)

	if err = cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}

func getContainerPidByName(containerName string) (string, error) {
	// 拼接出记录容器信息的文件路径
	dirURL := fmt.Sprintf(InfoLocFormat, containerName)
	configFilePath := dirURL + ConfigName
	// 读取内容并解析
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
