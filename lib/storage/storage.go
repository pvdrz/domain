package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
    "github.com/pvdrz/domain/lib/doc"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

func nextID(bucket *bolt.Bucket) (doc.DocumentID, error) {
    var id doc.DocumentID

    index, err := bucket.NextSequence()
    if err != nil {
        return id, err
    }

    binary.BigEndian.PutUint64(id[:], index)

    return id, nil
}

func OpenStorage(path string) (Storage, error) {
	var storage Storage

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return storage, err
	}

	storage.db = db

	err = storage.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("documents"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("hashes"))
		return err
	})

	return storage, err
}

func (storage *Storage) Insert(document *doc.Document) (doc.DocumentID, error) {
	var id doc.DocumentID

	err := storage.db.Update(func(tx *bolt.Tx) error {
		hash := document.Hash[:]

		hashes := tx.Bucket([]byte("hashes"))
		if hashes.Get(hash) != nil {
			return errors.New("duplicated hash")
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

func (storage *Storage) Get(id doc.DocumentID) (doc.Document, error) {
	var document doc.Document

	err := storage.db.View(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte("documents"))

        bytesDoc := documents.Get(id[:])
		buf := bytes.NewBuffer(bytesDoc)
		dec := gob.NewDecoder(buf)
		return dec.Decode(&document)

	})

	return document, err
}

func (storage *Storage) ForEach(f func(doc.DocumentID, doc.Document) error) error {
	return storage.db.View(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte("documents"))

		return documents.ForEach(func(bytesID []byte, bytesDoc []byte) error {
			var id doc.DocumentID
            copy(id[:], bytesID)

			var document doc.Document
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
