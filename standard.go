package log

import "log"

// Implements standard functions of go's log

// Printf log.Printf
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Println log.Println
func Println(v ...interface{}) {
	log.Println(v...)
}

// Add more when needed
