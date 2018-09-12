package witness

import (
	logger "log"
	"os"
	"time"
	"strconv"
)

const (
	TFORMAT = "2006-01-02 15H"
	LOG_FILE_MAX_LINE = 10000
	LOG_DIR = "logs/"
	LOG_EXT = ".log"
)

var log_count = 0
var show_log_console = true


type Logset struct {
	FileName string
	File *os.File
	Line int
}


func (l *Logset) Write(p []byte) (n int, err error) {
	if (l.Line > LOG_FILE_MAX_LINE) {
		setLoggerFile(l)
	}
	if show_log_console {
		os.Stdout.Write(p)
	}
	return l.File.Write(p)
}


func ShowLogConsole(s bool) {
	show_log_console = s
}


func (l *Logset) Close() {
	l.File.Close()
}


func log(a ...interface{}) {
	logger.Println(a...)
}


func Log(a ...interface{}) {
	log(a...)
}


func setLoggerFile(ls *Logset) {
	if ls.File != nil {
		ls.Close()
		ls.File = nil
	}

	os.Mkdir(LOG_DIR, 600)

	ls.FileName = LOG_DIR + time.Now().Format(TFORMAT);
	if (log_count > 0) {
		ls.FileName += "."+ strconv.Itoa(log_count)
	}
	ls.FileName += LOG_EXT
	log_count++

	file, err := os.OpenFile(ls.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		log("Cannot open log file", ls.FileName)
		return
	} 

	ls.File = file
	logger.SetOutput(ls)
}