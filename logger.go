package log

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
)

// Handle type
type Handle struct {
	logger     *log.Logger
	fileHandle *os.File
	Path       string
	Name       string
	Level      DebugLevel
	lock       sync.Mutex
}

func (handle *Handle) printf(level DebugLevel, format string, input ...interface{}) {
	if handle.logger == nil || handle.Level < level {
		return
	}

	extraFormat := ""
	if level == ERROR {
		_, file, line, _ := runtime.Caller(2)
		extraFormat = fmt.Sprintf("%s:%d ERROR ", file, line)
	} else if level == INFO {
		extraFormat = "INFO "
	} else if level == DEBUG {
		extraFormat = "DEBUG "
	}

	handle.lock.Lock()
	handle.logger.Printf(extraFormat+format, input...)
	handle.lock.Unlock()
}

// Println always adds <input> line to log file
func (handle *Handle) Println(input string) {
	if handle.logger == nil {
		return
	}

	handle.lock.Lock()
	handle.logger.Println(input)
	handle.lock.Unlock()
}

// PrintRestCall adds line to auditlog file with "<ip>: <action> - <result>"" format
func (handle *Handle) PrintRestCall(req *http.Request, action string, result string) {
	if handle.logger == nil {
		return
	}
	handle.lock.Lock()
	handle.logger.Printf("%s: %s - %s\n", req.RemoteAddr, action, result)
	handle.lock.Unlock()
}

// Error puts a string in the log with the ERROR debuglevel. Outputs file and linenumber where error occured.
func (handle *Handle) Error(input string) {
	handle.printf(ERROR, "%s\n", input)
}

// Errorf puts a format+string in the log with the ERROR debuglevel. Outputs file and linenumber where error occured.
func (handle *Handle) Errorf(format string, input ...interface{}) {
	handle.printf(ERROR, format, input...)
}

// Info puts a string in the log with the INFO debuglevel.
func (handle *Handle) Info(input string) {
	handle.printf(INFO, "%s\n", input)
}

// Infof puts a string in the log with the INFO debuglevel.
func (handle *Handle) Infof(format string, input ...interface{}) {
	handle.printf(INFO, format, input...)
}

// Debug puts a string in the log with the DEBUG debuglevel.
func (handle *Handle) Debug(input string) {
	handle.printf(DEBUG, "%s\n", input)
}

// Debugf puts a string in the log with the DEBUG debuglevel.
func (handle *Handle) Debugf(format string, input ...interface{}) {
	handle.printf(DEBUG, format, input...)
}
