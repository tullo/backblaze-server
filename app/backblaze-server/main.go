package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gabriel-vasile/mimetype"
	cached "github.com/tullo/backblaze-server/internal/cache"
	"github.com/tullo/conf"
)

type app struct {
	cache  *cached.Cache
	logger *log.Logger
	api    string
	keyID  string
	appKey string
	bucket string
	s3     string
	region string
	token  string
}

func main() {
	log := log.New(os.Stdout, "B2SERVER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	var cfg struct {
		Backblaze struct {
			ApplicationKey string `conf:"noprint"`
			KeyID          string `conf:"noprint"`
			Token          string `conf:"noprint"`
			API            string `conf:"default:https://api003.backblazeb2.com"`
			Endpoint       string `conf:"default:https://s3.eu-central-003.backblazeb2.com"`
			Region         string `conf:"default:eu-central-003"`
		}
		Environment string `conf:"default:development"`
		Domain      string `conf:"default:localhost"`
		Port        string `conf:"default::9090"`
	}

	if err := conf.Parse(os.Args[1:], "B2SERVER", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("B2SERVER", &cfg)
			if err != nil {
				panic(err)
			}
			fmt.Println(usage)
			return
		}
		panic(err)
	}

	out, err := conf.String(&cfg)
	if err != nil {
		panic(err)
	}
	log.Printf("configuration:\n%v\n", out)

	// RAM cache
	cache := cached.New()
	defer cache.Close()

	b2 := &cfg.Backblaze
	app := &app{
		cache:  cache,
		logger: log,
		api:    b2.API,
		keyID:  b2.KeyID,
		appKey: b2.ApplicationKey,
		s3:     b2.Endpoint,
		region: b2.Region,
		bucket: "",
		token:  "",
	}

	acc, err := authorizeAccount(app)
	if err != nil {
		log.Fatalf("failed to authorize: %v", err)
	}

	app.api = acc.APIURL // e.g https://api003.backblazeb2.com
	app.bucket = acc.Allowed.BucketName
	app.token = acc.AuthorizationToken

	http.HandleFunc("/", handler(app))

	switch cfg.Environment {
	case "development":
		log.Printf("serving files at http://%v%v\n", cfg.Domain, cfg.Port)
		panic(http.ListenAndServe(cfg.Port, nil))
	default:
		panic("Environment not set")
	}
}

func handler(app *app) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.RequestURI == "/favicon.ico" {
			return
		}

		objectKey := r.RequestURI[1:]

		if len(objectKey) == 0 {
			w.Write([]byte("serving files from backblaze"))
			return
		}

		app.logger.Println("serving object key", objectKey)

		// load from in-memory cache
		b, err := app.cache.Get([]byte(objectKey))
		if errors.Is(err, badger.ErrKeyNotFound) {
			// retrieve the file from backblaze
			res, err := retrieve(app, objectKey)
			if err != nil {
				app.logger.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				app.logger.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			if res.ContentType == nil || len(*res.ContentType) == 0 {
				mime := mimetype.Detect(b)
				w.Header().Set("Content-Type", mime.String())
				app.logger.Println("detected:", mime.String(), mime.Extension())
			}
			w.Header().Set("Content-Type", *res.ContentType)
			//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fn))
			io.Copy(w, bytes.NewReader(b))

			// cache the file
			f := cached.File{
				Body:     b,
				MimeType: *res.ContentType,
				FileName: objectKey,
			}
			err = app.cache.Set(f)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				app.logger.Fatal(err)
			}

			app.logger.Println("cached file:", objectKey)
			return
		}

		// serve the cached file
		var f cached.File
		if err = json.Unmarshal(b, &f); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			app.logger.Fatal(err)
		}

		app.logger.Println("using cached file:", f.FileName)
		w.Header().Set("Content-Length", strconv.Itoa(len(f.Body)))
		mime := mimetype.Detect(f.Body)
		w.Header().Set("Content-Type", mime.String())
		app.logger.Println("detected:", mime.String(), mime.Extension())
		io.Copy(w, bytes.NewReader(f.Body))
	}
}
