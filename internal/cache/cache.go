package cache

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v2"
)

// File represents an entry in the cache.
type File struct {
	Body     []byte
	MimeType string
	FileName string
}

// Cache implements a caching API.
type Cache struct {
	db *badger.DB
}

// New creates an instance of an in-memory database.
func New() Cache {
	var c Cache
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	c.db = db
	return c
}

// Close closes the underlying DB.
func (c Cache) Close() error {
	return c.db.Close()
}

// Get queries the cache for the specified key.
func (c Cache) Get(key []byte) ([]byte, error) {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	var val []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

// Set saves a file in the cache.
func (c Cache) Set(f File) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(f.FileName), b)
	})
}
