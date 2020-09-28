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

	"github.com/dgraph-io/badger/v2"
	"github.com/gabriel-vasile/mimetype"
	"github.com/packago/config"
	"github.com/tullo/backblaze-server/internal/cache"
)

func main() {

	log := log.New(os.Stdout, "BACKBLAZE : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	c := cache.New()
	defer c.Close()

	acc, err := authorizeAccount()
	if err != nil {
		log.Fatalf("failed to authorize: %v", err)
	}

	http.HandleFunc("/", handler(log, c, acc.Allowed.BucketName))

	switch config.File().GetString("environment") {
	case "development":
		addr := config.File().GetString("port.development")
		log.Println("serving files at", addr)
		panic(http.ListenAndServe(addr, nil))
	default:
		panic("Environment not set")
	}
}

func handler(log *log.Logger, c cache.Cache, bucketName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

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

		b, err := c.Get([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			// retrieve the file from backblaze
			res, err := serve(log, bucketName, key)
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
			f := cache.File{
				Body:     b,
				MimeType: *res.ContentType,
				FileName: key,
			}
			err = c.Set(f)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("cached file:", key)
			return
		}

		// serve the cached file
		var f cache.File
		if err = json.Unmarshal(b, &f); err != nil {
			log.Fatal(err)
		}

		log.Println("using cached file:", f.FileName)
		w.Header().Set("Content-Length", strconv.Itoa(len(f.Body)))
		mime := mimetype.Detect(f.Body)
		w.Header().Set("Content-Type", mime.String())
		log.Println("detected:", mime.String(), mime.Extension())
		io.Copy(w, bytes.NewReader(f.Body))
	}
}
