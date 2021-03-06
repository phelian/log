package log

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
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
	Verbose    bool
}

// SetFlags log.SetFlags, use to override default flags
func (handle *Handle) SetFlags(flag int) {
	handle.lock.Lock()
	defer handle.lock.Unlock()
	if handle.logger == nil {
		return
	}
	handle.logger.SetFlags(flag)
}

// SetPrefix log.SetPrefix, use to override default prefix
func (handle *Handle) SetPrefix(prefix string) {
	handle.lock.Lock()
	defer handle.lock.Unlock()
	if handle.logger == nil {
		return
	}
	handle.logger.SetPrefix(prefix)
}

func (handle *Handle) printf(level DebugLevel, format string, input ...interface{}) {
	extraFormat := ""
	if level == ERROR {
		_, file, line, _ := runtime.Caller(2)
		extraFormat = fmt.Sprintf("%s:%d ERROR ", path.Base(file), line)
	} else if level == INFO {
		extraFormat = "INFO "
	} else if level == DEBUG {
		extraFormat = "DEBUG "
	}

	handle.lock.Lock()
	defer handle.lock.Unlock()
	if handle.logger == nil || handle.Level < level {
		return
	}

	handle.logger.Printf(extraFormat+format, input...)
	if handle.Verbose {
		log.Printf(extraFormat+format, input...)
	}
}

// Println always adds <input> line to log file
func (handle *Handle) Println(input string) {
	handle.lock.Lock()
	defer handle.lock.Unlock()

	if handle.logger == nil {
		return
	}

	handle.logger.Println(input)

	if handle.Verbose {
		log.Println(input)
	}
}

// PrintRestCall adds line to auditlog file with "<ip>: <action> - <result>"" format
func (handle *Handle) PrintRestCall(req *http.Request, action string, result string) {
	handle.lock.Lock()
	defer handle.lock.Unlock()

	if handle.logger == nil {
		return
	}
	handle.logger.Printf("%s: %s - %s\n", req.RemoteAddr, action, result)

	if handle.Verbose {
		log.Printf("%s: %s - %s\n", req.RemoteAddr, action, result)
	}
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
