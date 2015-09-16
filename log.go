package log

import (
	"errors"
	"log"
	"os"
)

// DebugLevel for containing fixed levels
type DebugLevel int64

// The allowed loglevels
const (
	NONE  DebugLevel = -10
	ERROR DebugLevel = 0
	INFO  DebugLevel = 10
	DEBUG DebugLevel = 20
)

// Rotation configuration
type Rotation struct {
	Rotate           bool  `json:"rotate"`
	DaysBeforeRotate int64 `json:"days,omitempty"` // Number of days before rotating
	Size             int64 `json:"size,omitempty"` // Number of bytes of file before rotating
	DaysKeep         int64 `json:"keep,omitempty"` // Number of days to keep before removing
	MaxFilesKeep     int64 `json:"max_files_keep"` // Number of rotated logfiles to keep
	Compress         int64 `json:"compress"`       // Number of rotated files to keep before compressing, -1 = don't compress
}

// Config defines config attributes for auditlog module
type Config struct {
	Path     string   `json:"path`
	Name     string   `json:"name"`
	Level    string   `json:"level"`
	Rotation Rotation `json:"rotation"`
}

// New returns a new handle to log
func New(config Config) (*Handle, error) {
	file, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open [%s] for log", config.Path)
		return nil, errors.New("Failed to open path for log")
	}

	level := ERROR
	if config.Level == "INFO" {
		level = INFO
	} else if config.Level == "DEBUG" {
		level = DEBUG
	}

	handle := &Handle{
		logger:     log.New(file, config.Name+": ", log.Ldate|log.Ltime),
		fileHandle: file,
		Name:       config.Name,
		Level:      level,
		Path:       config.Path,
	}

	if config.Rotation.Rotate {
		err := setupRotation(handle, config.Rotation)
		if err != nil {
			return nil, err
		}
		handle.Println("Started")
	} else {
		handle.Println("Started, no log rotation selected")
	}

	return handle, nil
}
