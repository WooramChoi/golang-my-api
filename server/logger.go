package server

import (
	"io"
	"log"
	"os"
)

func GetLogger() *log.Logger {
	fpLog, err := os.OpenFile("logging.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)

	multiWriter := io.MultiWriter(os.Stdout, fpLog)
	logger.SetOutput(multiWriter)
	return logger
}
