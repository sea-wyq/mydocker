package container

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"mydocker/cgroups/subsystems"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/run/mydocker/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
)

type Info struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
}

func ListContainers() {
	// 读取存放容器信息目录下的所有文件
	files, err := os.ReadDir(InfoLoc)
	if err != nil {
		log.Errorf("read dir %s error %v", InfoLoc, err)
		return
	}
	containers := make([]*Info, 0, len(files))
	for _, file := range files {
		tmpContainer, err := GetContainerInfo(file)
		if err != nil {
			log.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 使用tabwriter.NewWriter在控制台打印出容器信息
	// tabwriter 是引用的text/tabwriter类库，用于在控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		log.Errorf("Fprint error %v", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			log.Errorf("Fprint error %v", err)
		}
	}
	if err = w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
	}
}

func GetContainerInfo(file fs.DirEntry) (*Info, error) {
	// 根据文件名拼接出完整路径
	containerName := file.Name()
	configFileDir := fmt.Sprintf(InfoLocFormat, containerName)
	configFileDir = configFileDir + ConfigName
	// 读取容器配置文件
	content, err := os.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	info := new(Info)
	if err = json.Unmarshal(content, info); err != nil {
		log.Errorf("json unmarshal error %v", err)
		return nil, err
	}

	return info, nil
}

func RecordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	// 随机出10位containerID
	id := RandStringBytes(IDLength)
	// 以当前时间作为容器创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	// 如果未指定容器名，则使用随机生成的containerID
	if containerName == "" {
		containerName = id
	}
	command := strings.Join(commandArray, "")
	containerInfo := &Info{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Name:        containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)
	// 拼接出存储容器信息文件的路径，如果目录不存在则级联创建
	dirUrl := fmt.Sprintf(InfoLocFormat, containerName)
	if err := os.MkdirAll(dirUrl, subsystems.Perm0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	// 将容器信息写入文件
	fileName := dirUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func DeleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(InfoLocFormat, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

func RandStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
