package main

import (
	"encoding/json"
	"io/ioutil"
    "github.com/pvdrz/domain/lib/doc"
)

type backup struct {
    Docs []doc.Document
};

func LoadBackup(path string) ([]doc.Document, error) {
    var backup backup

    bytes, err := ioutil.ReadFile(path)

    if err != nil {
        return nil, err;
    }

    err = json.Unmarshal(bytes, &backup)

    if err != nil {
        return nil, err;
    }

    return backup.Docs, nil
}

func DumpBackup(path string, docs []doc.Document) error {
    bytes, err := json.Marshal(backup { Docs: docs })

    if err != nil {
        return err
    }

    return ioutil.WriteFile(path, bytes, 0664)
}
