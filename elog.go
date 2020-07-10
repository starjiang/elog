package elog

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LOG_LEVEL_DEBUG         = 1
	LOG_LEVEL_INFO          = 2
	LOG_LEVEL_WARN          = 3
	LOG_LEVEL_ERROR         = 4
	LOG_LEVEL_NONE          = 5
	LOG_MAX_FILE_SIZE       = 1024 * 1024 * 1024
	LOG_MAX_BUFFER_SIZE     = 1024 * 1024
	LOG_MAX_ROTATE_FILE_NUM = 10
	LOG_DEPTH_GLOBAL        = 4
	LOG_DEPTH_HANDLER       = 3
)

func init() {
	flag.BoolVar(&logger.logToStderr, "logToStderr", false, "log to stderr,default false")
	flag.IntVar(&logger.flushTime, "logFlushTime", 3, "log flush time interval,default 3 seconds")
	flag.StringVar(&logger.logLevel, "logLevel", "INFO", "log level[DEBUG,INFO,WARN,ERROR,NONE],default INFO level")
	flag.StringVar(&logger.logPath, "logPath", "", "log path,default log to current directory")
	logger.depth = LOG_DEPTH_GLOBAL
	go logger.flushDaemon()
}

type EasyLogger struct {
	mutex       sync.Mutex
	logToStderr bool
	flushTime   int
	logLevel    string
	writer      EasyLogHandler
	depth       int
	logPath     string
}

func NewEasyLogger(logLevel string, logToStderr bool, flushTime int, writer EasyLogHandler) *EasyLogger {

	logger := &EasyLogger{}
	logger.logLevel = logLevel
	logger.logToStderr = logToStderr
	logger.flushTime = flushTime
	logger.writer = writer
	logger.depth = LOG_DEPTH_HANDLER
	go logger.flushDaemon()
	return logger
}

type EasyLogHandler interface {
	io.Writer
	Flush()
}

func NewEasyFileHandler(path string, bufferSize int) *EasyFileHandler {
	handler := &EasyFileHandler{}
	handler.path = path
	handler.file = nil
	handler.buffer = nil
	handler.currentDate = ""
	handler.bufferSize = bufferSize
	return handler
}

type EasyFileHandler struct {
	path        string
	file        *os.File
	buffer      *bufio.Writer
	bufferSize  int
	currentDate string
	nbytes      int
}

func (efh *EasyFileHandler) Write(data []byte) (int, error) {

	err := efh.rotateFile()

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return 0, err
	}
	efh.nbytes += len(data)
	return efh.buffer.Write(data)

}

func (efh *EasyFileHandler) Flush() {
	if efh.file != nil {
		efh.buffer.Flush()
		//efh.file.Sync()
	}
}

func (efh *EasyFileHandler) rotateFile() error {

	var err error
	date := getTimeNowDate()

	if efh.path == "" {
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		efh.path = path
	}
	if efh.currentDate != date {
		if efh.file != nil {
			efh.buffer.Flush()
			err = efh.file.Close()
			if err != nil {
				return err
			}
			efh.file = nil
		}
		efh.currentDate = date
	}

	if efh.nbytes > LOG_MAX_FILE_SIZE {
		appName := getAppName()
		efh.buffer.Flush()
		err = efh.file.Close()
		if err != nil {
			return err
		}

		efh.file = nil

		logFilePath := efh.path + string(os.PathSeparator) + appName + "-" + date + ".log." + strconv.Itoa(LOG_MAX_ROTATE_FILE_NUM-1)
		if fileIsExist(logFilePath) {
			err = os.Remove(logFilePath)
			if err != nil {
				return err
			}
		}

		for i := LOG_MAX_ROTATE_FILE_NUM - 2; i >= 0; i-- {
			var logFilePath string
			if i == 0 {
				logFilePath = efh.path + string(os.PathSeparator) + appName + "-" + date + ".log"
			} else {
				logFilePath = efh.path + string(os.PathSeparator) + appName + "-" + date + ".log." + strconv.Itoa(i)
			}
			if fileIsExist(logFilePath) {
				logFileNewPath := efh.path + string(os.PathSeparator) + appName + "-" + date + ".log." + strconv.Itoa(i+1)
				err := os.Rename(logFilePath, logFileNewPath)
				if err != nil {
					return err
				}
			}
		}
	}

	if efh.file == nil {
		logFilePath := efh.path + string(os.PathSeparator) + getAppName() + "-" + date + ".log"
		efh.file, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		efh.nbytes = 0
		efh.buffer = bufio.NewWriterSize(efh.file, efh.bufferSize)
	}
	return nil
}

func getLogLevelInt(level string) int {
	if level == "DEBUG" {
		return LOG_LEVEL_DEBUG
	} else if level == "INFO" {
		return LOG_LEVEL_INFO
	} else if level == "WARN" {
		return LOG_LEVEL_WARN
	} else if level == "ERROR" {
		return LOG_LEVEL_ERROR
	} else if level == "NONE" {
		return LOG_LEVEL_NONE
	}
	return LOG_LEVEL_INFO
}

func getLogLevelString(level int) string {
	if level == LOG_LEVEL_DEBUG {
		return "DEBUG"
	} else if level == LOG_LEVEL_INFO {
		return "INFO"
	} else if level == LOG_LEVEL_WARN {
		return "WARN"
	} else if level == LOG_LEVEL_ERROR {
		return "ERROR"
	} else if level == LOG_LEVEL_NONE {
		return "NONE"
	}
	return "INFO"
}

func getAppName() string {
	appName := os.Args[0]
	slash := strings.LastIndex(appName, string(os.PathSeparator))
	if slash >= 0 {
		appName = appName[slash+1:]
	}
	return appName
}

func (el *EasyLogger) getHeader(level int, writer io.Writer) {

	_, file, line, ok := runtime.Caller(el.depth)

	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	fmt.Fprintf(writer, "[%s][%s][file:%s line:%d] ", getLogLevelString(level), getTimeNowStr(), file, line)
	if el.logToStderr {
		fmt.Fprintf(os.Stderr, "[%s][%s][file:%s line:%d] ", getLogLevelString(level), getTimeNowStr(), file, line)
	}
}

func (el *EasyLogger) output(level int, args ...interface{}) {

	if el.depth == LOG_DEPTH_GLOBAL && !flag.Parsed() {
		os.Stderr.Write([]byte("ERROR: logging before flag.Parse\n"))
		return
	}
	if level < getLogLevelInt(el.logLevel) {
		return
	}
	el.mutex.Lock()
	defer el.mutex.Unlock()
	if el.writer == nil {
		el.writer = NewEasyFileHandler(el.logPath, LOG_MAX_BUFFER_SIZE) //delay init
	}
	el.getHeader(level, el.writer)
	fmt.Fprintln(el.writer, args...)
	if el.logToStderr {
		fmt.Fprintln(os.Stderr, args...)
	}
}

func (el *EasyLogger) outputf(level int, format string, args ...interface{}) {

	if el.depth == LOG_DEPTH_GLOBAL && !flag.Parsed() {
		os.Stderr.Write([]byte("ERROR: logging before flag.Parse\n"))
		return
	}
	if level < getLogLevelInt(el.logLevel) {
		return
	}

	el.mutex.Lock()
	defer el.mutex.Unlock()

	if el.writer == nil {
		el.writer = NewEasyFileHandler(el.logPath, LOG_MAX_BUFFER_SIZE) //delay init
	}

	el.getHeader(level, el.writer)
	fmt.Fprintf(el.writer, format, args...)
	el.writer.Write([]byte("\n"))
	if el.logToStderr {
		fmt.Fprintf(os.Stderr, format, args...)
		os.Stderr.WriteString("\n")
	}
}

func (el *EasyLogger) Flush() {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	if el.writer != nil {
		el.writer.Flush()
	}
}

func (el *EasyLogger) Debug(args ...interface{}) {
	el.output(LOG_LEVEL_DEBUG, args...)
}
func (el *EasyLogger) Debugf(format string, args ...interface{}) {
	el.outputf(LOG_LEVEL_DEBUG, format, args...)
}

func (el *EasyLogger) Info(args ...interface{}) {
	el.output(LOG_LEVEL_INFO, args...)
}
func (el *EasyLogger) Infof(format string, args ...interface{}) {
	el.outputf(LOG_LEVEL_INFO, format, args...)
}

func (el *EasyLogger) Warn(args ...interface{}) {
	el.output(LOG_LEVEL_WARN, args...)
}
func (el *EasyLogger) Warnf(format string, args ...interface{}) {
	el.outputf(LOG_LEVEL_WARN, format, args...)
}

func (el *EasyLogger) Error(args ...interface{}) {
	el.output(LOG_LEVEL_ERROR, args...)
}

func (el *EasyLogger) Errorf(format string, args ...interface{}) {
	el.outputf(LOG_LEVEL_ERROR, format, args...)
}

func (el *EasyLogger) Println(args ...interface{}) {
	el.output(LOG_LEVEL_INFO, args...)
}

func (el *EasyLogger) Printf(format string, args ...interface{}) {
	el.outputf(LOG_LEVEL_INFO, format, args...)
}

func (el *EasyLogger) flushDaemon() {
	for _ = range time.NewTicker(time.Second * time.Duration(el.flushTime)).C {
		el.Flush()
	}
}

var logger EasyLogger

func Debug(args ...interface{}) {
	logger.Debug(args...)
}
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)

}
func Warningf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Println(args ...interface{}) {
	logger.Println(args...)
}
func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Flush() {
	logger.Flush()
}

func getTimeNow() int64 {
	return time.Now().UnixNano() / 1e6
}

func getTimeNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getTimeNowDate() string {
	return time.Now().Format("2006-01-02")
}

func fileIsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
