package db

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"

	"github.com/pvdrz/domain/lib/doc"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	inner *bolt.DB
}

func nextID(bucket *bolt.Bucket) (doc.DocID, error) {
	var id doc.DocID

	index, err := bucket.NextSequence()
	if err != nil {
		return id, err
	}

	binary.BigEndian.PutUint64(id[:], index)

	return id, nil
}

func OpenDB(path string) (DB, error) {
	var db DB

	inner, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return db, err
	}

	db.inner = inner

	err = db.inner.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("documents"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("hashes"))
		return err
	})

	return db, err
}

func (db *DB) Insert(document *doc.Doc) (doc.DocID, error) {
	var id doc.DocID

	err := db.inner.Update(func(tx *bolt.Tx) error {
		hash := document.Hash[:]

		hashes := tx.Bucket([]byte("hashes"))
		if hashes.Get(hash) != nil {
			sHash := hex.EncodeToString(hash)
			return fmt.Errorf("the document with title \"%s\" cannot be inserted because the hash \"%s\" is already in the database", document.Title, sHash)
		}

		documents := tx.Bucket([]byte("documents"))
		newID, err := nextID(documents)
		if err != nil {
			return err
		}

		id = newID

		err = hashes.Put(hash, id[:])
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(document)
		if err != nil {
			return err
		}

		return documents.Put(id[:], buf.Bytes())
	})

	return id, err
}

func (db *DB) Get(id doc.DocID) (doc.Doc, error) {
	var document doc.Doc

	err := db.inner.View(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte("documents"))

		bytesDoc := documents.Get(id[:])
		buf := bytes.NewBuffer(bytesDoc)
		dec := gob.NewDecoder(buf)
		return dec.Decode(&document)

	})

	return document, err
}

func (db *DB) Delete(id doc.DocID) error {
	return db.inner.Update(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte("documents"))

		return documents.Delete(id[:])
	})
}

func (db *DB) ForEach(f func(doc.DocID, doc.Doc) error) error {
	return db.inner.View(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte("documents"))

		return documents.ForEach(func(bytesID []byte, bytesDoc []byte) error {
			var id doc.DocID
			copy(id[:], bytesID)

			var document doc.Doc
			buf := bytes.NewBuffer(bytesDoc)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&document)
			if err != nil {
				return err
			}

			return f(id, document)
		})
	})
}
