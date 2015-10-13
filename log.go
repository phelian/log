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

// Config defines config attributes for auditlog module
type Config struct {
	Path     string         `json:"path`
	Name     string         `json:"name.omitempty"`
	Level    string         `json:"level,omitempty"`
	Verbose  bool           `json:"verbose"`
	Rotate   bool           `json:"rotate"`
	Rotation RotationConfig `json:"rotation"`
}

// New returns a new handle to log
func New(config Config) (*Handle, error) {
	file, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open [%s] for log\n", config.Path)
		return nil, errors.New("Failed to open path for log")
	}

	level := ERROR
	if config.Level == "INFO" {
		level = INFO
	} else if config.Level == "DEBUG" {
		level = DEBUG
	}

	prefix := ""
	if config.Name != "" {
		prefix = config.Name + ": "
	}

	handle := &Handle{
		logger:     log.New(file, prefix, log.Ldate|log.Ltime),
		fileHandle: file,
		Name:       config.Name,
		Level:      level,
		Path:       config.Path,
		Verbose:    config.Verbose,
	}

	if config.Rotate {
		err := handle.SetupRotation(config.Rotation, config.Verbose)
		if err != nil {
			return nil, err
		}
		handle.Println("Started")
	} else {
		handle.Println("Started, no log rotation selected")
	}

	return handle, nil
}
