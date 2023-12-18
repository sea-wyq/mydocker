package container

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	LogFile = "container.log"
)

func LogContainer(containerName string) {
	logFileLocation := fmt.Sprintf(InfoLocFormat, containerName) + LogFile
	file, err := os.Open(logFileLocation)
	if err != nil {
		log.Errorf("log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("log container read file %s error %v", logFileLocation, err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		log.Errorf("log container Fprint  error %v", err)
		return
	}
	defer file.Close()
}
