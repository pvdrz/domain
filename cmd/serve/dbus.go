package main

import (
	"encoding/hex"
	"errors"
	"path"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/pvdrz/domain/lib/doc"
	"github.com/skratchdot/open-golang/open"
)

const serverName = "com.github.pvdrz.domain"
const serverPath = "/com/github/pvdrz/domain"

func ServeDbus(domain Domain) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Export(&domain, serverPath, "org.gnome.Shell.SearchProvider2")

	reply, err := conn.RequestName(serverName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}

	select {}
}

func (domain *Domain) GetInitialResultSet(terms []string) ([]string, *dbus.Error) {
	query := strings.Join(terms, " ")
	ids := domain.search(query)

	results := make([]string, len(ids))
	for i, id := range ids {
		results[i] = id.ToString()
	}

	return results, nil
}

func (domain *Domain) GetSubsearchResultSet(_ []string, terms []string) ([]string, *dbus.Error) {
	return domain.GetInitialResultSet(terms)
}

func (domain *Domain) GetResultMetas(strIDs []string) ([]map[string]dbus.Variant, *dbus.Error) {
	metas := make([]map[string]dbus.Variant, len(strIDs))

	for i, strID := range strIDs {
		id, err := doc.DocIDFromString(strID)
		if err != nil {
			return metas, dbus.MakeFailedError(err)
		}

		doc, err := domain.get(id)
		if err != nil {
			return metas, dbus.MakeFailedError(err)
		}

		meta := make(map[string]dbus.Variant, 3)
		meta["id"] = dbus.MakeVariant(strID)
		meta["name"] = dbus.MakeVariant(doc.Title)
		meta["description"] = dbus.MakeVariant(strings.Join(doc.Authors, ", "))

		metas[i] = meta
	}

	return metas, nil
}

func (domain *Domain) ActivateResult(strID string, _ []string, _ uint32) *dbus.Error {
	id, err := doc.DocIDFromString(strID)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	doc, err := domain.get(id)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	filename := hex.EncodeToString(doc.Hash[:]) + "." + doc.Extension
	path := path.Join(domain.config.Path, filename)
	err = open.Start(path)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	return nil
}

func (domain *Domain) LaunchSearch(_ []string, _ uint32) *dbus.Error {
	return nil
}
