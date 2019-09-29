package base

import (
	"io/ioutil"
	"log"
	"os"
)

func LogToTemp() func() {
	logFile, err := ioutil.TempFile("/tmp", "leap")
	if err != nil {
		log.Panic("leap couldn't make a logging file: %v", err)
	}
	log.SetOutput(logFile)

	return func() {
		log.SetOutput(os.Stderr)
	}
}
