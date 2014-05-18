package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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
	md5hex   string
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

	h := md5.New()
	_, _ = h.Write(data)

	s.assets[file.Name()] = &asset{
		md5hex:   hex.EncodeToString(h.Sum(nil)),
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

		asset, ok := static.Get(path)
		if !ok {
			http.NotFound(w, r)
			log.Printf("[INFO] Not found.")
			return
		}

		for _, etag := range r.Header["If-None-Match"] {
			if etag == asset.md5hex {
				w.WriteHeader(http.StatusNotModified)
				log.Printf("[Info] Not modified")
				return
			}
		}

		w.Header().Set("ETag", asset.md5hex)
		w.Header().Set("Content-Type", asset.mimetype)
		io.Copy(w, asset.Reader())

	}, err
}
