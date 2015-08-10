package log

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL

	Ldate         = log.Ldate         // the date: 2013/08/23
	Ltime         = log.Ltime         // the time: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
)

type Logger struct {
	level    int
	seq      int
	filename string
	msg      chan *logMsg
	logfile  *os.File
	log.Logger
}

type logMsg struct {
	level int
	msg   string
}

var std = NewLogger("log"+string(os.PathSeparator)+"service.log", "", Ldate|Ltime)

func NewLogger(filename string, prefix string, flag int) (newLogger *Logger) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln("Open log file failed:", err)
	}
	l := &Logger{DEBUG, 1, filename, make(chan *logMsg, 10), f, *log.New(f, prefix, flag)}
	l.StartLogger()
	return l
}

func (logger *Logger) writerMsg(loglevel int, msg string) {
	if logger.level > loglevel {
		return
	}
	lm := new(logMsg)
	lm.level = loglevel
	lm.msg = msg
	logger.msg <- lm
	return
}

func (logger *Logger) SetLevel(l int) {
	logger.level = l
}

func (logger *Logger) refreshLogfile() {

	stat, err := logger.logfile.Stat()
	if err != nil {
		log.Fatalln("Read log file error")
	}

	if stat.Size() > 1500000 {
		prefix := logger.Logger.Prefix()
		flag := logger.Logger.Flags()
		logger.logfile.Close()
		os.Rename(logger.filename, logger.filename+"."+strconv.Itoa(logger.seq))

		f, err := os.OpenFile(logger.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln("Open log file failed: ", err)
		}
		logger.logfile = f

		logger.Logger = *log.New(f, prefix, flag)
		logger.seq++
	}
}
func (logger *Logger) loopwrite() {
	for {
		select {
		case bm := <-logger.msg:
			logger.Output(2, bm.msg)
			if bm.level == FATAL {
				p, _ := os.FindProcess(os.Getgid())
				p.Signal(syscall.SIGINT)
			}
		}
	}
}

func (logger *Logger) StartLogger() {
	go logger.loopwrite()
	var refreshfunc func()
	refreshfunc = func() {
		logger.refreshLogfile()
		time.AfterFunc(300*time.Second, refreshfunc)
	}
	refreshfunc()
}

func (logger *Logger) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf("[DEBUG] "+format, v...)
	logger.writerMsg(DEBUG, msg)
}

func (logger *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf("[INFO] "+format, v...)
	logger.writerMsg(INFO, msg)
}

func (logger *Logger) Warning(format string, v ...interface{}) {
	msg := fmt.Sprintf("[WARNING] "+format, v...)
	logger.writerMsg(WARNING, msg)
}
func (logger *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	logger.writerMsg(ERROR, msg)
}

func (logger *Logger) Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf("[FATAL] "+format, v...)
	logger.writerMsg(FATAL, msg)
}

func (logger *Logger) Close() {
	logger.logfile.Close()
}

func Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf("[DEBUG] "+format, v...)
	std.writerMsg(DEBUG, msg)
}

func Info(format string, v ...interface{}) {
	msg := fmt.Sprintf("[INFO] "+format, v...)
	std.writerMsg(INFO, msg)
}

func Warning(format string, v ...interface{}) {
	msg := fmt.Sprintf("[WARNING] "+format, v...)
	std.writerMsg(WARNING, msg)
}

func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	std.writerMsg(ERROR, msg)
}

func Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf("[FATAL] "+format, v...)
	std.writerMsg(FATAL, msg)
}

func Close() {
	std.Close()
}
