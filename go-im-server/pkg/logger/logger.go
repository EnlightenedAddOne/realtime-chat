package logger

import (
	"log"
	"os"
	"time"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func Init() {
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func GetLogFileName() string {
	return time.Now().Format("2006-01-02") + ".log"
}

func Infof(format string, v ...interface{}) {
	Info.Printf(format, v...)
}
