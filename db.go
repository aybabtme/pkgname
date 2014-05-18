package main

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var queueSize = 100

var nameSources = []string{
	"seed/names.flatfile",
}

type Filter func(string) error

type DB struct {
	lock    sync.RWMutex
	names   []string
	r       *rand.Rand
	filters []Filter

	goods *leakingQueue
	bads  *leakingQueue
}

func NewDB() *DB {

	db := &DB{
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
		goods: newQueue(queueSize),
		bads:  newQueue(queueSize),
		filters: []Filter{
			noHyphens,
			noUnderscore,
			notCapitalized,
			noReferenceToGo,
			noReferenceToGolang,
			validPackageNames,
		},
	}

	var goodNames []string
	for _, name := range loadNames(nameSources) {
		errs := db.Validate(name)
		if len(errs) != 0 {
			log.Printf("[DB] Rejecting %q from source: \n%s", name, strings.Join(errs, "\n"))
		} else {
			goodNames = append(goodNames, name)
		}
	}
	db.names = goodNames
	lengthFilter, mean, stdev := closeToMean(goodNames)
	log.Printf("[DB] Mean name length=%f, stdev=%f.", mean, stdev)
	db.filters = append(db.filters, lengthFilter)

	return db
}

func (db *DB) Get() string {
	db.lock.RLock()
	defer db.lock.RUnlock()
	max := len(db.names)
	index := db.r.Intn(max)
	return db.names[index]
}

func (db *DB) Validate(pkgname string) []string {
	db.lock.RLock()
	var errStrs []string
	for _, filter := range db.filters {
		err := filter(pkgname)
		if err != nil {
			errStrs = append(errStrs, err.Error())
		}
	}
	db.lock.RUnlock()

	db.lock.Lock()
	if len(errStrs) == 0 {
		db.goods.Enqueue(pkgname)
	} else {
		db.bads.Enqueue(pkgname)
	}
	db.lock.Unlock()

	return errStrs
}

func (db *DB) Last(last int) ([]string, []string) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.goods.Last(last), db.bads.Last(last)
}

func loadNames(sources []string) []string {
	var names []string
	for _, filename := range sources {
		names = append(names, loadSource(filename)...)
	}
	return names
}

func loadSource(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("opening %q: %v", filename, err)
	}
	defer func() { _ = file.Close() }()

	var r io.Reader
	if filepath.Ext(filename) == ".gz" {
		reader, err := gzip.NewReader(file)
		if err != nil {
			log.Fatalf("[ERROR] %q is not a valid GZIP file: %v", filename, err)
		}
		defer func() { _ = reader.Close() }()
		r = reader
	} else {
		r = file
	}

	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanLines)

	var names []string

	for scan.Scan() {
		names = append(names, strings.TrimSpace(scan.Text()))
	}

	if err := scan.Err(); err != nil {
		log.Fatalf("scanning %q: %v", filename, err)
	}

	return names
}
