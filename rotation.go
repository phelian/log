package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// RotationConfig ...
type RotationConfig struct {
	DaysBeforeRotate int64 `json:"days,omitempty"` // Number of days before rotating
	Size             int64 `json:"size,omitempty"` // Number of bytes of file before rotating
	DaysKeep         int64 `json:"keep,omitempty"` // Number of days to keep before removing
	MaxFilesKeep     int64 `json:"max_files_keep"` // Number of rotated logfiles to keep
	Compress         int64 `json:"compress"`       // Number of rotated files to keep before compressing, -1 = don't compress
}

// SetupRotation starts a rotation on handle based on rotation configuration
func (handle *Handle) SetupRotation(config RotationConfig, verbose bool) error {
	// Rotation Trigger #1 Size
	if config.Size > 0 {
		if verbose {
			log.Printf("Starting logrotate service for file %s with byte trigger of %d bytes.", handle.Path, config.Size)
		}
		handle.printf(INFO, "Log will rotate after about %d bytes", config.Size)

		go func() {
			for {
				fileInfo, err := os.Stat(handle.Path)
				if err == nil {
					if fileInfo.Size() > config.Size {
						handle.rotate(verbose)
						handle.analyseLogFiles(config, verbose)
					}
				}
				time.Sleep(time.Duration(1) * time.Minute)
			}
		}()
	}

	// Rotation Trigger #2 Age
	if config.DaysBeforeRotate > 0 {
		if verbose {
			log.Printf("Starting logrotate service for file %s with day trigger of %d days.", handle.Path, config.DaysBeforeRotate)
		}
		handle.Infof("Log will rotate after %d days", config.DaysBeforeRotate)
		go func(seconds time.Duration) {
			for {
				time.Sleep(seconds)
				handle.rotate(verbose)
				handle.analyseLogFiles(config, verbose)
			}
		}(time.Duration(config.DaysBeforeRotate) * time.Hour * 24)
	}

	if verbose {
		if config.DaysKeep > 0 {
			log.Printf("Rotated files will be removed after %d days.", config.DaysKeep)
		}

		if config.MaxFilesKeep > 0 {
			log.Printf("Max %d rotated files will be kept.", config.MaxFilesKeep)
		}

		if config.Compress > -1 {
			log.Printf("Will start to compress after %d rotations.", config.Compress)
		}
	}

	// Possible Configuration Issue
	if config.DaysKeep <= 0 && config.MaxFilesKeep <= 0 {
		fmt.Printf("!! WARNING !!\nLogs will be kept forever, this is probably not what you intended. Please set keep and/or max_files_keep for each log entry in config\n!! WARNING !!\n")
	}

	// Analyse log files at startup
	handle.analyseLogFiles(config, verbose)
	return nil
}

func (handle *Handle) rotate(verbose bool) error {
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
	rotatedName := handle.Path + "." + time.Now().Format(time.RFC3339)
	if err == nil {
		err = os.Rename(handle.Path, rotatedName)
		if err != nil {
			return err
		}
	}

	if handle.fileHandle != nil {
		handle.fileHandle, err = os.OpenFile(handle.Path, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Println("Failed to open [%s] for log", handle.Path)
			return errors.New("Failed to open path for log")
		}

		handle.logger = log.New(handle.fileHandle, handle.Name+": ", log.Ldate|log.Ltime)
	}

	if verbose {
		log.Println("Rotated file", handle.Path, "into", rotatedName)
	}

	return nil
}

// Removes log files older then <configurable> days, compresses after <configurable> days
// keep max <configurable> files
func (handle *Handle) analyseLogFiles(config RotationConfig, verbose bool) {
	var fileList []string

	pathDir := path.Dir(handle.Path) + "/"
	baseRotatedFilename := path.Base(handle.Path) + "."
	filepath.Walk(pathDir, processLogfile(baseRotatedFilename, config.DaysKeep, &fileList, verbose))
	numberOfFiles := int64(len(fileList))

	// Loop filelist and compress if specified and meets criteria
	if config.Compress > -1 {
		for i := int64(0); i < numberOfFiles-config.Compress; i++ {
			if strings.Contains(fileList[i], ".gz") {
				continue
			}
			gzipCompress(fileList[i])
			if verbose {
				log.Println("Gzipped", fileList[i])
			}
		}
	}

	// Remove oldest files in list if configured
	if config.MaxFilesKeep > 0 && numberOfFiles > config.MaxFilesKeep {
		for i := int64(0); i < numberOfFiles-config.MaxFilesKeep; i++ {
			os.Remove(fileList[i])
			if verbose {
				log.Println("Removed", fileList[i], "due to reached amount of logfiles.")
			}
		}
	}
}

// The files are walked in lexical order, which makes the output deterministic
func processLogfile(basefile string, daysKeep int64, fileList *[]string, verbose bool) filepath.WalkFunc {
	return func(fullpath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if !strings.Contains(fullpath, basefile) {
			return nil
		}

		fileTimeStamp := path.Base(fullpath)[len(basefile):]
		if strings.Contains(fileTimeStamp, ".gz") {
			fileTimeStamp = fileTimeStamp[:len(fileTimeStamp)-3]
		}

		logfileDate, err := time.Parse(time.RFC3339, fileTimeStamp)
		if err != nil {
			return err
		}

		*fileList = append(*fileList, fullpath)

		// Remove if too old
		if daysKeep > 0 {
			timePasedSinceCreation := time.Since(logfileDate)
			if int64(timePasedSinceCreation.Hours()) > daysKeep*24 {
				os.Remove(fullpath)
				if verbose {
					log.Println("Removed ", fullpath, "due to reached days to keep.")
				}
			}
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
