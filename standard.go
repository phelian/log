package log

import "log"

// Implements standard functions of go's log

// Printf calls go standard log.Printf
func Printf(format string, v ...interface{}) {
	log.Printf(format, v)
}

// Println calls go standard log.Println
func Println(v ...interface{}) {
	log.Println(v)
}

// Add more when needed
