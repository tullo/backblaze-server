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

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gabriel-vasile/mimetype"
	"github.com/packago/config"
	"github.com/tullo/bzserver/bzserver"
)

type File struct {
	Body     []byte
	MimeType string
	FileName string
}

func main() {

	log := log.New(os.Stdout, "BZSERVER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	opt := badger.DefaultOptions("").WithInMemory(true)
	cache, err := badger.Open(opt)
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	acc, err := bzserver.AuthorizeAccount()
	if err != nil {
		log.Fatalf("failed to authorize: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.RequestURI == "/favicon.ico" {
			return
		}

		fn := r.RequestURI[1:]
		// load from in-memory cache
		b, err := get(cache, []byte(fn))
		if errors.Is(err, badger.ErrKeyNotFound) {
			// retrieve the file from backblaze
			bucket := acc.Allowed.BucketName
			res, err := bzserver.Serve(log, bucket, fn)
			if err != nil {
				log.Fatal(err)
			}

			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
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
			err = set(cache, File{
				Body:     b,
				MimeType: *res.ContentType,
				FileName: fn,
			})
			if err != nil {
				log.Fatal(err)
			}

			log.Println("cached file:", fn)
			return
		}

		// serve the cached file
		var f File
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

	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

func set(db *badger.DB, f File) error {
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
