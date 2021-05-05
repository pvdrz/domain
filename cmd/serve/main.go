package main

import (
    "path"
	config "github.com/pvdrz/domain"
	"github.com/pvdrz/domain/lib/doc"
	"github.com/pvdrz/domain/lib/storage"
	"github.com/pvdrz/domain/lib/text"
)

func main() {
	domain, err := NewDomain()
	if err != nil {
		panic(err)
	}

	ServeDbus(domain)
}

type Domain struct {
	storage storage.Storage
	index   text.Index
	config  config.Config
}

func NewDomain() (Domain, error) {
	var domain Domain

	config, err := config.OpenConfig()
	if err != nil {
		return domain, err
	}

	storagePath := path.Join(config.Path, "db")
	storage, err := storage.OpenStorage(storagePath)
	if err != nil {
		return domain, err
	}

	index := text.NewIndex()
	err = storage.ForEach(func(id doc.DocID, doc doc.Doc) error {
		index.Insert(id, &doc)
		return nil
	})
	if err != nil {
		return domain, err
	}

	domain.storage = storage
	domain.index = index
	domain.config = config

	return domain, err
}

func (domain *Domain) Get(id doc.DocID) (doc.Doc, error) {
	return domain.storage.Get(id)
}

func (domain *Domain) Search(query string) []doc.DocID {
	return domain.index.Search([]byte(query))
}

func (domain *Domain) Insert(document *doc.Doc) (doc.DocID, error) {
	id, err := domain.storage.Insert(document)
	if err != nil {
		return id, err
	}

	domain.index.Insert(id, document)

	return id, nil
}
