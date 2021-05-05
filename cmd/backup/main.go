package main

import (
	"encoding/json"
	"github.com/pvdrz/domain/lib/doc"
	"io/ioutil"
)

type backup struct {
	Docs []doc.Doc
}

func loadBackup(path string) ([]doc.Doc, error) {
	var backup backup

	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &backup)

	if err != nil {
		return nil, err
	}

	return backup.Docs, nil
}

func dumpBackup(path string, docs []doc.Doc) error {
	bytes, err := json.Marshal(backup{Docs: docs})

	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0664)
}
