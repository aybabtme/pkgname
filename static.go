package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type asset struct {
	mimetype string
	content  []byte
}

func (a *asset) Reader() io.Reader {
	return bytes.NewBuffer(a.content)
}

type staticDB struct {
	basePath string
	assets   map[string]*asset
	lock     sync.RWMutex
}

func (s *staticDB) Put(file *os.File, size int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.assets[file.Name()]; ok {
		return fmt.Errorf("file already known, %q", file.Name())
	}

	buf := bytes.NewBuffer(nil)

	_, err := io.Copy(buf, file)
	if err != nil {
		return err
	}

	data := buf.Bytes()

	mimetype := mime.TypeByExtension(filepath.Ext(file.Name()))
	if mimetype == "" {
		log.Printf("[INFO] Couldn't detect mimetype from exntension, sniffing content: %q", file.Name())
		mimetype = http.DetectContentType(data)
	}
	log.Printf("[INFO] Mimetype of %q: %q", file.Name(), mimetype)

	s.assets[file.Name()] = &asset{
		content:  data,
		mimetype: mimetype,
	}

	return nil
}

func (s *staticDB) Get(path string) (*asset, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	a, ok := s.assets[path]
	return a, ok
}

func staticHandler(path string) (http.HandlerFunc, error) {
	static := &staticDB{
		basePath: path,
		assets:   make(map[string]*asset),
	}

	err := filepath.Walk(path, func(name string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return err
		}

		log.Printf("[INFO] StaticDB: Considering %q", name)

		file, err := os.Open(name)
		if err != nil {
			return fmt.Errorf("opening file for static DB, %v", err)
		}

		if err := static.Put(file, int(fi.Size())); err != nil {
			return fmt.Errorf("putting file in static DB, %v", err)
		}

		log.Printf("[INFO] StaticDB: Added to DB, %q", name)

		return err
	})

	return func(w http.ResponseWriter, r *http.Request) {

		var path string
		if r.URL.Path == "/" {
			path = static.basePath + "index.html"
		} else {
			path = static.basePath + r.URL.Path[1:]
		}

		log.Printf("[INFO] Request for assets %q", path)

		asset, ok := static.Get(path)
		if !ok {
			http.NotFound(w, r)
			log.Printf("[INFO] Not found.")
			return
		}
		w.Header().Set("Content-Type", asset.mimetype)
		io.Copy(w, asset.Reader())
	}, err
}
