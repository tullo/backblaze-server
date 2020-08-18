package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/packago/config"
	"github.com/tullo/bzserver/bzserver"
)

func main() {
	acc, err := bzserver.AuthorizeAccount()
	if err != nil {
		log.Fatalf("failed to authorize: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		log.Println("handling:", r.RequestURI)

		if r.RequestURI == "/favicon.ico" {
			return
		}

		bucket := acc.Allowed.BucketName
		res, err := bzserver.Serve(bucket, r.RequestURI[1:])
		if err != nil {
			fmt.Println(err)
		}

		b, err := ioutil.ReadAll(res.Body)
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
		//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", r.RequestURI[1:]))
		io.Copy(w, bytes.NewReader(b))
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
