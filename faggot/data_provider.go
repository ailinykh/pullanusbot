package faggot

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
)

var mutexMap = map[string]*sync.Mutex{}

// One more mutex to prevent concurrent map writes
var mutex sync.Mutex

// DataProvider safe reads and writes to files
type DataProvider struct {
	WorkingDir string
}

// NewDataProvider func is a DataProvider constructor
func NewDataProvider(args ...string) *DataProvider {
	dp := DataProvider{WorkingDir: "data"}

	if len(args) > 0 {
		dp.WorkingDir = args[0]
	}

	if _, err := os.Stat(dp.WorkingDir); os.IsNotExist(err) {
		log.Printf("DATA: Directory not exist! Creating directory: %s", dp.WorkingDir)
		err = os.MkdirAll(dp.WorkingDir, os.ModePerm)
		if err != nil {
			log.Fatalf("DATA: Can't create directory: %s", dp.WorkingDir)
		}
	}

	log.Printf("DATA: Using directory: %s", dp.WorkingDir)

	return &dp
}

func (d *DataProvider) saveJSON(filename string, data []byte) {
	mutex.Lock()
	_, ok := mutexMap[filename]
	if !ok {
		mutexMap[filename] = &sync.Mutex{}
	}
	mutexMap[filename].Lock()
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		mutexMap[filename].Unlock()
		mutex.Unlock()
	}()

	regexp := regexp.MustCompile(`[-\d]+`)
	prefix := regexp.FindString(filename)
	log.Printf("%s DATA: Saving game Mutex unlocked (%s)", prefix, filename)

	file := path.Join(d.WorkingDir, filename)
	err := ioutil.WriteFile(file, data, 0644)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s DATA: JSON saved (%s)", prefix, filename)
}

func (d *DataProvider) loadJSON(filename string) []byte {
	mutex.Lock()
	_, ok := mutexMap[filename]
	if !ok {
		mutexMap[filename] = &sync.Mutex{}
	}
	mutexMap[filename].Lock()
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		mutexMap[filename].Unlock()
		mutex.Unlock()
	}()

	regexp := regexp.MustCompile(`[-\d]+`)
	prefix := regexp.FindString(filename)
	log.Printf("%s DATA: Loading json Mutex unlocked (%s)", prefix, filename)

	file := path.Join(d.WorkingDir, filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Printf("%s DATA: File not found! Trying to create... (%s)", prefix, filename)
		ioutil.WriteFile(file, []byte("{}"), 0644)
	}

	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s DATA: JSON loaded (%s)", prefix, filename)
	return data
}
