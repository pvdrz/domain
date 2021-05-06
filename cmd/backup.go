package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pvdrz/domain/lib/db"
	"github.com/pvdrz/domain/lib/doc"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:                   "load [path]",
	Short:                 "Load a backup",
	Long:                  "Load the contents of a json file to the database. Be sure that the server is not running before executing this command",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := openConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		db, err := db.OpenDB(conf.pathDB())
		if err != nil {

			fmt.Println(err)
			return
		}

		docs, err := loadBackup(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, doc := range docs {
			_, err = db.Insert(&doc)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	},
}

var saveCmd = &cobra.Command{
	Use:                   "save [path]",
	Short:                 "Save a backup",
	Long:                  "Save the contents the database into a json file. Be sure that the server is not running before executing this command",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := openConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		db, err := db.OpenDB(conf.pathDB())
		if err != nil {
			fmt.Println(err)
			return
		}

		docs := make([]doc.Doc, 0)

		err = db.ForEach(func(_ doc.DocID, document doc.Doc) error {
			docs = append(docs, document)
			return nil
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		err = saveBackup(args[0], docs)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

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

func saveBackup(path string, docs []doc.Doc) error {
	bytes, err := json.Marshal(backup{Docs: docs})
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0664)
}
