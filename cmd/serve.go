package cmd

import (
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/pvdrz/domain/lib/db"
	"github.com/pvdrz/domain/lib/doc"
	"github.com/pvdrz/domain/lib/text"

	"github.com/godbus/dbus/v5"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the domain dbus server",
	Long:  "Run the dbus server that can be used as a search provider for gnome",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

const serverName = "com.github.pvdrz.domain"
const serverPath = "/com/github/pvdrz/domain"

type server struct {
	db     db.DB
	index  text.Index
	config config
}

func serve() error {
	server, err := newServer()
	if err != nil {
		return err
	}

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Export(&server, serverPath, "org.gnome.Shell.SearchProvider2")

	reply, err := conn.RequestName(serverName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("the server name \"%s\" is already taken", serverName)
	}

	select {}
}

func newServer() (server, error) {
	var server server

	config, err := openConfig()
	if err != nil {
		return server, err
	}

	db, err := db.OpenDB(config.pathDB())
	if err != nil {
		return server, err
	}

	index := text.NewIndex()
	err = db.ForEach(func(id doc.DocID, doc doc.Doc) error {
		index.Insert(id, &doc)
		return nil
	})
	if err != nil {
		return server, err
	}

	server.db = db
	server.index = index
	server.config = config

	return server, err
}

func (server *server) get(id doc.DocID) (doc.Doc, error) {
	return server.db.Get(id)
}

func (server *server) search(query string) []doc.DocID {
	return server.index.Search([]byte(query))
}

func (server *server) insert(document *doc.Doc) (doc.DocID, error) {
	id, err := server.db.Insert(document)
	if err != nil {
		return id, err
	}

	server.index.Insert(id, document)

	return id, nil
}

func (server *server) GetInitialResultSet(terms []string) ([]string, *dbus.Error) {
	query := strings.Join(terms, " ")
	ids := server.search(query)

	results := make([]string, len(ids))
	for i, id := range ids {
		results[i] = id.ToString()
	}

	return results, nil
}

func (server *server) GetSubsearchResultSet(_ []string, terms []string) ([]string, *dbus.Error) {
	return server.GetInitialResultSet(terms)
}

func (server *server) GetResultMetas(identifiers []string) ([]map[string]dbus.Variant, *dbus.Error) {
	metas := make([]map[string]dbus.Variant, len(identifiers))

	for pos, identifier := range identifiers {
		id, err := doc.DocIDFromString(identifier)
		if err != nil {
			return metas, dbus.MakeFailedError(err)
		}

		doc, err := server.get(id)
		if err != nil {
			return metas, dbus.MakeFailedError(err)
		}

		meta := make(map[string]dbus.Variant, 3)
		meta["id"] = dbus.MakeVariant(identifier)
		meta["name"] = dbus.MakeVariant(doc.Title)
		meta["description"] = dbus.MakeVariant(strings.Join(doc.Authors, ", "))

		metas[pos] = meta
	}

	return metas, nil
}

func (server *server) ActivateResult(identifier string, _ []string, _ uint32) *dbus.Error {
	id, err := doc.DocIDFromString(identifier)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	doc, err := server.get(id)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	filename := hex.EncodeToString(doc.Hash[:]) + "." + doc.Extension
	path := path.Join(server.config.Path, filename)
	err = open.Start(path)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	return nil
}

func (server *server) LaunchSearch(_ []string, _ uint32) *dbus.Error {
	return nil
}
