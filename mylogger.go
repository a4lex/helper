package helper

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type myLogger struct {
	file    *os.File
	mu      *sync.Mutex
	verbose int
}

const (
	FATAL int = 1 << iota
	ERROR
	INFO
	FUNC
	DEBUG
	CUSTOM
)

var (
	logInstance *myLogger
	initLogOnce sync.Once

	strErrorVerbose = map[int]string{
		FATAL:  `FATAL`,
		ERROR:  `ERROR`,
		INFO:   `INFO`,
		FUNC:   `FUNC`,
		DEBUG:  `DEBUG`,
		CUSTOM: `CUSTOM`,
	}
)

func LogInit(logFile string, logLevel int) error {
	var f *os.File
	var err error = fmt.Errorf("log already inited")

	initLogOnce.Do(func() {
		f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err == nil {
			logInstance = &myLogger{f, &sync.Mutex{}, logLevel}
		}
	})

	return err
}

func LogRelease() {
	if logInstance != nil {
		logInstance.verbose = 0
		logInstance.file.Close()
	}
}

func Fatal(str string, v ...interface{}) { log(FATAL, str, v...) }
func Error(str string, v ...interface{}) { log(ERROR, str, v...) }
func Info(str string, v ...interface{})  { log(INFO, str, v...) }
func Func(str string, v ...interface{})  { log(FUNC, str, v...) }
func Debug(str string, v ...interface{}) { log(DEBUG, str, v...) }

func CustomLogFunc(level int, strDesc string) func(string, ...interface{}) {
	return func(str string, v ...interface{}) {
		// important do like that because if we close file in LogRelease logInstance.verbose = 0
		if level&logInstance.verbose == level {

			output := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), strDesc, fmt.Sprintf(str, v...))
			fmt.Fprint(os.Stdout, output)

			logInstance.mu.Lock()
			fmt.Fprint(logInstance.file, output)
			logInstance.mu.Unlock()
		}
	}

}

func log(level int, str string, v ...interface{}) {
	if level&logInstance.verbose == level {

		output := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), strErrorVerbose[level], fmt.Sprintf(str, v...))
		fmt.Fprint(os.Stdout, output)

		logInstance.mu.Lock()
		fmt.Fprint(logInstance.file, output)
		logInstance.mu.Unlock()
	}

	if level == FATAL {
		os.Exit(1)
	}
}
