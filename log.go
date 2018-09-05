package main

import (
	logger "log"
	"os"
	"time"
)

const TFORMAT = "2006-01-02 15.04"
const LOG_FILE_MAX_LINE = 10000

type Logset struct {
	FileName string
	File *os.File
	Line int
}


func (l *Logset) Write(p []byte) (n int, err error) {
	if (l.Line > LOG_FILE_MAX_LINE) {
		setLoggerFile(l)
	}
	os.Stdout.Write(p)
	return l.File.Write(p)
}


func (l *Logset) Close() {
	l.File.Close()
}


func log(a ...interface{}) {
	logger.Println(a...)
}


func setLoggerFile(ls *Logset) {
	if ls.File != nil {
		ls.Close()
		ls.File = nil
	}

	ls.FileName = time.Now().Format(TFORMAT) + ".log"
	file, err := os.OpenFile(ls.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		log("Cannot open log file", ls.FileName)
		return
	} 

	ls.File = file
	logger.SetOutput(ls)
}