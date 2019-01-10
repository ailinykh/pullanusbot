package faggot

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
)

var dataDir = "data"
var mtxs = map[string]*sync.Mutex{}

// One more mutex to prevent concurrent map writes
var mutex = sync.Mutex{}

// DataProvider safe reads and writes to files
type DataProvider struct {
}

func (d *DataProvider) saveJSON(filename string, data []byte) {
	mtx, ok := mtxs[filename]
	if !ok {
		mtx = &sync.Mutex{}
		mutex.Lock()
		mtxs[filename] = mtx
		mutex.Unlock()
	}
	mtx.Lock()
	defer mtx.Unlock()

	regexp := regexp.MustCompile(`[-\d]+`)
	prefix := regexp.FindString(filename)
	log.Printf("%s DP:   Saving game Mutex unlocked (%s)", prefix, filename)

	file := path.Join(dataDir, filename)
	err := ioutil.WriteFile(file, data, 0644)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s DP:   JSON saved (%s)", prefix, filename)
}

func (d *DataProvider) loadJSON(filename string) []byte {
	mtx, ok := mtxs[filename]
	if !ok {
		mtx = &sync.Mutex{}
		mutex.Lock()
		mtxs[filename] = mtx
		mutex.Unlock()
	}
	mtx.Lock()
	defer mtx.Unlock()

	regexp := regexp.MustCompile(`[-\d]+`)
	prefix := regexp.FindString(filename)
	log.Printf("%s DP:   Loading json Mutex unlocked (%s)", prefix, filename)

	file := path.Join(dataDir, filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Printf("%s DP:   File not found! Trying to create... (%s)", prefix, filename)
		ioutil.WriteFile(file, []byte("{}"), 0644)
	}

	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s DP:   JSON loaded (%s)", prefix, filename)
	return data
}
