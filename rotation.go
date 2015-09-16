package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (handle *Handle) rotate() error {
	handle.lock.Lock()
	defer handle.lock.Unlock()

	// Close existing file if open
	if handle.fileHandle != nil {
		err := handle.fileHandle.Close()
		handle.fileHandle = nil
		if err != nil {
			return err
		}
	}

	// Rename dest file
	_, err := os.Stat(handle.Path)
	if err == nil {
		err = os.Rename(handle.Path, handle.Path+"."+time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	handle.fileHandle, err = os.OpenFile(handle.Path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("Failed to open [%s] for log", handle.Path)
		return errors.New("Failed to open path for log")
	}

	handle.logger = log.New(handle.fileHandle, handle.Name+": ", log.Ldate|log.Ltime)

	return nil
}

func setupRotation(handle *Handle, config Rotation) error {
	if config.Size > 0 {
		handle.printf(INFO, "Log will rotate after about %d bytes", config.Size)

		go func() {
			for {
				fileInfo, err := os.Stat(handle.Path)
				if err == nil {
					if fileInfo.Size() > config.Size {
						handle.rotate()
					}
				}
				time.Sleep(time.Duration(1) * time.Hour)
			}
		}()
	}

	if config.DaysBeforeRotate > 0 {
		handle.printf(INFO, "Log will rotate after %d days", config.DaysBeforeRotate)
		go func(seconds time.Duration) {
			for {
				time.Sleep(seconds)
				handle.rotate()
			}
		}(time.Duration(config.DaysBeforeRotate) * time.Hour * 24)
	}

	if config.DaysKeep <= 0 {
		log.Println("Logs will be kept forever, this is probably not what you intended. Please set rotated_keep for each log entry in config")
		handle.Println("Wrong config, missing rotated_keep value for log")
		panic("I do not want to run under these conditions!")
	}

	// Analyse log files at startup and twice every day
	go func() {
		for {
			analyseLogFiles(handle, config.DaysKeep)
			time.Sleep(time.Duration(1) * time.Hour * 12)
		}
	}()
	return nil
}

// Removes log files older then <configurable> days, compresses after two days
func analyseLogFiles(handle *Handle, daysToKeep int64) {
	path := handle.Path[:strings.LastIndex(handle.Path, "/")+1]
	baseRotatedFilename := handle.Path[strings.LastIndex(handle.Path, "/")+1:] + "."
	filepath.Walk(path, processLogfile(baseRotatedFilename, daysToKeep))
}

func processLogfile(basefile string, daysToKeep int64) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if !strings.Contains(path, basefile) {
			return nil
		}

		fileTimeStamp := path[len(basefile):]
		if strings.Contains(fileTimeStamp, ".gz") {
			fileTimeStamp = fileTimeStamp[:len(fileTimeStamp)-3]
		}

		logfileDate, err := time.Parse(time.RFC3339, fileTimeStamp)
		if err != nil {
			return err
		}
		timePasedSinceCreation := time.Since(logfileDate)

		// Remove if too old
		if int64(timePasedSinceCreation.Hours()) > daysToKeep*24 {
			os.Remove(path)
		} else if int64(timePasedSinceCreation.Hours()) > 3*24 {
			gzipCompress(path)
		}

		return nil
	}
}

func gzipCompress(path string) error {
	rawfile, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rawfile.Close()

	// calculate the buffer size for rawfile
	info, _ := rawfile.Stat()
	size := info.Size()
	rawbytes := make([]byte, size)

	// read rawfile content into buffer
	buffer := bufio.NewReader(rawfile)
	_, err = buffer.Read(rawbytes)
	if err != nil {
		log.Println(err)
		return err
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	writer.Write(rawbytes)
	writer.Close()

	err = ioutil.WriteFile(path+".gz", buf.Bytes(), 0666)
	if err != nil {
		log.Println(err)
		return err
	}

	os.Remove(path)
	return err
}
