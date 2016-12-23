package adbutil

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
)

type PerfLogger struct {
	LogFilePath   string
	FileHandle    *os.File
	FileLogger    *log.Logger
	StdoutLogger  *log.Logger
	PrintCodeLine bool
	Locker        *sync.Mutex
}

func NewPerfLogger(logFilePath string) *PerfLogger {
	var logFile *os.File
	var fileErr error
	logFile, fileErr = os.OpenFile(logFilePath, os.O_RDONLY, 0)
	if fileErr != nil && os.IsNotExist(fileErr) {
		// 日志文件不存在 创建
		logFile, fileErr = os.OpenFile(logFilePath, os.O_CREATE, 0666)
	} else {
		// 存在 追加
		logFile, fileErr = os.OpenFile(logFilePath, os.O_APPEND, 0666)
	}

	fileLogger := log.New(logFile, "", log.Ldate|log.Ltime)
	stdoutLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return &PerfLogger{LogFilePath: logFilePath, FileHandle: logFile, FileLogger: fileLogger, StdoutLogger: stdoutLogger}
}

// 开启后，由于锁的存在，导致并发性能受到严重影响，只适用于调试模式
func (pl *PerfLogger) EnablePrintCodeLine() {
	pl.Locker = &sync.Mutex{}
	pl.PrintCodeLine = true
}

func (pl *PerfLogger) info(content string, prefix string, a ...interface{}) {
	if pl.PrintCodeLine {
		pl.Locker.Lock()
		_, file, line, _ := runtime.Caller(2)
		prefix = prefix + fmt.Sprintf(" %s %d ", file, line)
		defer pl.Locker.Unlock()
	}

	pl.FileLogger.SetPrefix(prefix)
	pl.StdoutLogger.SetPrefix(prefix)
	pl.FileLogger.Printf(content, a...)
	pl.StdoutLogger.Printf(content, a...)
}

func (pl *PerfLogger) Info(content string, a ...interface{}) {
	pl.info(content, "[INFO] ", a...)
}

func (pl *PerfLogger) Debug(content string, a ...interface{}) {
	pl.info(content, "[DEBUG] ", a...)
}

func (pl *PerfLogger) Warn(content string, a ...interface{}) {
	pl.info(content, "[WARN] ", a...)
}

func (pl *PerfLogger) Error(content string, a ...interface{}) {
	pl.info(content, "[ERROR] ", a...)
}

func (pl *PerfLogger) SetPrefix(prefix string) {
	pl.FileLogger.SetPrefix(prefix)
	pl.StdoutLogger.SetPrefix(prefix)
}

func (pl *PerfLogger) Close() {
	pl.FileHandle.Close()
}
