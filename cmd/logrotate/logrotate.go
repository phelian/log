package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/phelian/log"
)

// Standalone logrotater
var verbose bool
var path string
var configPath string
var size int64

func main() {
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Usage = usage

	// Multiple files options
	flag.StringVar(&configPath, "c", "", "Config location")

	// Single file options
	flag.StringVar(&path, "p", "", "Path to single file")
	flag.Int64Var(&size, "s", 0, "Size before rotation of single file")
	maxFiles := flag.Int64("m", 5, "Max files to keep for single file")
	compress := flag.Int64("z", 0, "Number of files rotated before compressing, -1 does not compress")

	flag.Parse()

	// Start handler(s)
	if configPath != "" {
		var configs map[string]log.RotationConfig
		err := readConfig(configPath, &configs)
		if err != nil {
			fmt.Println("Error: ", err.Error())
			return
		}

		for filePath, fileConfig := range configs {
			handle := &log.Handle{Path: filePath}
			handle.SetupRotation(fileConfig, verbose)
		}
	} else if path != "" || size != 0 {
		config := log.RotationConfig{Size: size, Compress: *compress, MaxFilesKeep: *maxFiles}
		handle := &log.Handle{Path: path}

		err := handle.SetupRotation(config, verbose)
		if err != nil {
			log.Println(err.Error())
			return
		}
	} else {
		usage()
		return
	}

	for {
		select {}
	}
}

func usage() {
	fmt.Println("logrotate: usage: [v] [c file ...|| ps [mz]]")
	fmt.Println("")
	fmt.Println("Global")
	fmt.Println("-v true/false	Optional	Verbose")
	fmt.Println(" 				Default: false")
	fmt.Println("Multi file mode:")
	fmt.Println("-c path		Path to config")
	fmt.Println("Single file mode:")
	fmt.Println("-p path		Mandatory 	Path to file")
	fmt.Println("-s size		Mandatory	Size in bytes to trigger rotation")
	fmt.Println("-m maxFiles	Optional	Max files to keep in total.")
	fmt.Println(" 				Default: 5")
	fmt.Println("-z compress	Optional	Number of files to keep uncompressed. -1 Turns off compression.")
	fmt.Println("				Default: 0")
	fmt.Println("")
}

func readConfig(path string, config *map[string]log.RotationConfig) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, config)
	if err != nil {
		return err
	}

	return err
}
