package main

import (
	"encoding/json"
	"flag"
	"github.com/aybabtme/httpgzip"
	"log"
	"net/http"
	"runtime"
	"strings"
	"text/template"
)

func main() {

	port := flag.String("port", "5000", "port to listen on")
	numCPU := flag.Int("cpu", runtime.NumCPU(), "number of cpus to use")
	dev := flag.Bool("dev", false, "dev mode uses a static file handler that reads from the FS at each request")

	flag.Parse()

	runtime.GOMAXPROCS(*numCPU)

	log.SetFlags(log.Flags() | log.Lshortfile | log.Lmicroseconds)

	db := NewDB()
	registerFilters(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", jsontype(validate(db)))
	mux.HandleFunc("/history", jsontype(history(db)))
	mux.HandleFunc("/generate", jsontype(generate(db)))

	if *dev {
		mux.Handle("/", http.FileServer(http.Dir("static/")))
	} else {
		static, err := staticHandler("static/")
		if err != nil {
			log.Fatalf("[ERROR] Failed to prepare static assets, %v", err)
		}
		mux.HandleFunc("/", static)
	}

	laddr := ":" + *port
	log.Printf("Listening on %q with %d cores", laddr, *numCPU)

	gziper := httpgzip.NewHandler(mux)

	if err := http.ListenAndServe(laddr, gziper); err != nil {
		log.Fatalf("[ERROR] Failed to listen and serve: %v", err)
	}
}

func jsontype(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f(w, r)
	}
}

func validate(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, `{"error": "Can only POST on this endpoint."}`, http.StatusTeapot)
			return
		}

		pkgname := r.FormValue("pkgname")
		if pkgname == "" {
			http.Error(w, `{"error": "Need a package name."}`, http.StatusBadRequest)
			return
		}

		errs := db.Validate(pkgname)

		data, err := json.Marshal(struct {
			Err     string   `json:"error"`
			Success bool     `json:"success"`
			Pkgname string   `json:"pkgname"`
			Causes  []string `json:"causes"`
		}{
			Err:     "",
			Success: len(errs) == 0,
			Pkgname: pkgname,
			Causes:  errs,
		})

		if err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Printf("[ERROR] Couldn't send validation to client: %v", err)
		}
	}
}

func history(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error": "Can only GET on this endpoint."}`, http.StatusTeapot)
			return
		}

		goods, bads := db.Last(10)

		data, err := json.Marshal(struct {
			Goods []string `json:"goods"`
			Bads  []string `json:"bads"`
		}{
			Goods: goods,
			Bads:  bads,
		})

		if err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Printf("[ERROR] Couldn't send list of bad names to client: %v", err)
		}
	}
}

func generate(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" {
			http.Error(w, `{"error": "Can only GET on this endpoint."}`, http.StatusTeapot)
			return
		}

		data, err := json.Marshal(struct {
			Err     string `json:"error"`
			Pkgname string `json:"pkgname"`
		}{
			Err:     "",
			Pkgname: db.Get(),
		})

		if err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Printf("[ERROR] Couldn't send generated name to client: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, err error) {
	titled := strings.Title(err.Error())
	escaped := template.JSEscapeString(titled)
	http.Error(w, `{"error": "`+escaped+`."}`, http.StatusInternalServerError)
}
