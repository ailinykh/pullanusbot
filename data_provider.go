package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
)

// Mutex to prevent concurrent file reading/writing
var mutex sync.Mutex

// DataProvider safe reads and writes to files
type DataProvider struct {
	WorkingDir string
}

// NewDataProvider func is a DataProvider constructor
func NewDataProvider(args ...string) (*DataProvider, error) {
	dp := DataProvider{WorkingDir: "data"}

	if len(args) > 0 {
		dp.WorkingDir = args[0]
	}

	if _, err := os.Stat(dp.WorkingDir); os.IsNotExist(err) {
		log.Printf("DATA: Directory not exist! Creating directory: %s", dp.WorkingDir)
		err = os.MkdirAll(dp.WorkingDir, os.ModePerm)
		if err != nil {
			log.Printf("DATA: Can't create directory: %s", dp.WorkingDir)
			return nil, err
		}
	}

	log.Printf("DATA: Using directory: %s", dp.WorkingDir)

	return &dp, nil
}

func (d *DataProvider) saveJSON(filename string, data []byte) error {
	mutex.Lock()
	defer mutex.Unlock()

	file := path.Join(d.WorkingDir, filename)
	return ioutil.WriteFile(file, data, 0644)
}

func (d *DataProvider) loadJSON(filename string) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()

	regexp := regexp.MustCompile(`[-\d]+`)
	prefix := regexp.FindString(filename)

	file := path.Join(d.WorkingDir, filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Printf("%s DATA: File not found! Trying to create... (%s)", prefix, filename)
		ioutil.WriteFile(file, []byte("{}"), 0644)
	}

	return ioutil.ReadFile(file)
}
