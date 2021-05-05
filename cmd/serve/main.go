package main

import (
    "path"
	config "github.com/pvdrz/domain"
	"github.com/pvdrz/domain/lib/doc"
	"github.com/pvdrz/domain/lib/db"
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
	db db.DB
	index   text.Index
	config  config.Config
}

func NewDomain() (Domain, error) {
	var domain Domain

	config, err := config.OpenConfig()
	if err != nil {
		return domain, err
	}

	dbPath := path.Join(config.Path, "db")
	db, err := db.OpenDB(dbPath)
	if err != nil {
		return domain, err
	}

	index := text.NewIndex()
	err = db.ForEach(func(id doc.DocID, doc doc.Doc) error {
		index.Insert(id, &doc)
		return nil
	})
	if err != nil {
		return domain, err
	}

	domain.db = db
	domain.index = index
	domain.config = config

	return domain, err
}

func (domain *Domain) get(id doc.DocID) (doc.Doc, error) {
	return domain.db.Get(id)
}

func (domain *Domain) search(query string) []doc.DocID {
	return domain.index.Search([]byte(query))
}

func (domain *Domain) insert(document *doc.Doc) (doc.DocID, error) {
	id, err := domain.db.Insert(document)
	if err != nil {
		return id, err
	}

	domain.index.Insert(id, document)

	return id, nil
}
