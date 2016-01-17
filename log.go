package gotrade

import (
	"github.com/Sirupsen/logrus"
	"os"
)

func NewLogger(name string) (logger *logrus.Logger) {
	logger = logrus.New()
	logOutput, err := os.OpenFile(GetBasePath()+"/log/"+name+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	logger.Out = logOutput
	logger.Formatter = new(logrus.JSONFormatter)
	return logger
}

func GetBasePath() string {
	// @todo small hack
	base, err := os.Getwd()
	if err != nil {
		return "."
	}
	if isExist(base) {
		return base
	} else if isExist(base + "/..") {
		return base + "/.."
	} else if isExist(base + "/../..") {
		return base + "/../.."
	} else if isExist(base + "/../../..") {
		return base + "/../../.."
	} else {
		return "."
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path + "/config")
	return err == nil || os.IsExist(err)
}
