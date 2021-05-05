package main

import (
	"encoding/hex"
	"errors"
	"path"
	"strings"

    "github.com/pvdrz/domain/lib/doc"
	"github.com/godbus/dbus/v5"
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

	methods := make(map[string]interface{}, 5)
	methods["GetInitialResultSet"] = func(terms []string) ([]string, *dbus.Error) {
		result, err := getInitialResultSet(&domain, terms)
		if err != nil {
			return result, dbus.MakeFailedError(err)
		}
		return result, nil
	}
	methods["GetSubsearchResultSet"] = func(previousResults []string, terms []string) ([]string, *dbus.Error) {
		result, err := getSubsearchResultSet(&domain, previousResults, terms)
		if err != nil {
			return result, dbus.MakeFailedError(err)
		}
		return result, nil
	}
	methods["GetResultMetas"] = func(identifiers []string) ([]map[string]dbus.Variant, *dbus.Error) {
		result, err := getResultMetas(&domain, identifiers)
		if err != nil {
			return result, dbus.MakeFailedError(err)
		}
		return result, nil
	}
	methods["ActivateResult"] = func(identifier string, terms []string, timestamp uint32) *dbus.Error {
		err := activateResult(&domain, identifier, terms, timestamp)
		if err != nil {
			return dbus.MakeFailedError(err)
		}
		return nil
	}
	methods["LaunchSearch"] = func(terms []string, timestamp uint32) *dbus.Error {
		err := launchSearch(&domain, terms, timestamp)
		if err != nil {
			return dbus.MakeFailedError(err)
		}
		return nil
	}
	conn.ExportMethodTable(methods, serverPath, "org.gnome.Shell.SearchProvider2")

	reply, err := conn.RequestName(serverName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}

	select {}
}

func getInitialResultSet(domain *Domain, terms []string) ([]string, error) {
	query := strings.Join(terms, " ")
	ids := domain.Search(query)

	results := make([]string, len(ids))
	for i, id := range ids {
		results[i] = id.ToString()
	}

	return results, nil
}

func getSubsearchResultSet(domain *Domain, _ []string, terms []string) ([]string, error) {
	return getInitialResultSet(domain, terms)
}

func getResultMetas(domain *Domain, strIDs []string) ([]map[string]dbus.Variant, error) {
	metas := make([]map[string]dbus.Variant, len(strIDs))

	for i, strID := range strIDs {
		id, err := doc.DocumentIDFromString(strID)
		if err != nil {
			return metas, err
		}

		doc, err := domain.Get(id)
		if err != nil {
			return metas, err
		}

		meta := make(map[string]dbus.Variant, 3)
		meta["id"] = dbus.MakeVariant(strID)
		meta["name"] = dbus.MakeVariant(doc.Title)
		meta["description"] = dbus.MakeVariant(strings.Join(doc.Authors, ", "))

		metas[i] = meta
	}

	return metas, nil
}

func activateResult(domain *Domain, strID string, _ []string, _ uint32) error {
	id, err := doc.DocumentIDFromString(strID)
	if err != nil {
		return err
	}

	doc, err := domain.Get(id)
	if err != nil {
		return err
	}

	filename := hex.EncodeToString(doc.Hash[:]) + "." + doc.Extension
	path := path.Join(domain.config.Path, filename)
	return open.Start(path)
}

func launchSearch(domain *Domain, _ []string, _ uint32) error {
	return nil
}
