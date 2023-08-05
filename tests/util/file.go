package util

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
)

func SaveErrorLogToFile(logName string, buf bytes.Buffer) {
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	if err = ioutil.WriteFile(logName, buf.Bytes(), 0644); err != nil {
		log.Printf("error save log to file:%v", logName)
	} else {
		log.Printf("save log to file:%v", logName)
	}
}
