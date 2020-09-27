package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/gabriel-vasile/mimetype"
	"github.com/packago/config"
)

type file struct {
	Body     []byte
	MimeType string
	FileName string
}

func main() {

	log := log.New(os.Stdout, "BACKBLAZE : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	opt := badger.DefaultOptions("").WithInMemory(true)
	cache, err := badger.Open(opt)
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	acc, err := authorizeAccount()
	if err != nil {
		log.Fatalf("failed to authorize: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.RequestURI == "/favicon.ico" {
			return
		}

		key := r.RequestURI[1:]

		if len(key) == 0 {
			w.Write([]byte("serving files from backblaze"))
			return
		}

		log.Println("serving file", key)

		// load from in-memory cache
		b, err := get(cache, []byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			// retrieve the file from backblaze
			bucket := acc.Allowed.BucketName
			res, err := serve(log, bucket, key)
			if err != nil {
				log.Println(err)
				return
			}

			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Println(err)
				return
			}

			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			if res.ContentType == nil || len(*res.ContentType) == 0 {
				mime := mimetype.Detect(b)
				w.Header().Set("Content-Type", mime.String())
				log.Println("detected:", mime.String(), mime.Extension())
			}
			w.Header().Set("Content-Type", *res.ContentType)
			//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fn))
			io.Copy(w, bytes.NewReader(b))

			// cache the file
			err = set(cache, file{
				Body:     b,
				MimeType: *res.ContentType,
				FileName: key,
			})
			if err != nil {
				log.Fatal(err)
			}

			log.Println("cached file:", key)
			return
		}

		// serve the cached file
		var f file
		if err = json.Unmarshal(b, &f); err != nil {
			log.Fatal(err)
		}

		log.Println("using cached file:", f.FileName)
		w.Header().Set("Content-Length", strconv.Itoa(len(f.Body)))
		mime := mimetype.Detect(f.Body)
		w.Header().Set("Content-Type", mime.String())
		log.Println("detected:", mime.String(), mime.Extension())
		io.Copy(w, bytes.NewReader(f.Body))
	})

	switch config.File().GetString("environment") {
	case "development":
		addr := config.File().GetString("port.development")
		log.Println("serving files at", addr)
		panic(http.ListenAndServe(addr, nil))
	default:
		panic("Environment not set")
	}
}

func get(db *badger.DB, key []byte) ([]byte, error) {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	var val []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

func set(db *badger.DB, f file) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(f.FileName), b)
	})
}
