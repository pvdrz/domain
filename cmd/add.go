package cmd

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a document",
	Long:  "Add a document to the domain DB",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		err := addDocumentReq(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var title string
var authors []string
var keywords []string

func init() {
	addCmd.Flags().StringVarP(&title, "title", "t", "", "Title of the document")
	addCmd.MarkFlagRequired("title")

	addCmd.Flags().StringSliceVarP(&authors, "authors", "a", make([]string, 0), "Authors of the document")
	addCmd.MarkFlagRequired("authors")

	addCmd.Flags().StringSliceVarP(&keywords, "keywords", "k", make([]string, 0), "Keywords of the document")
	addCmd.MarkFlagRequired("keywords")
}

func addDocumentReq(path string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Object(serverName, serverPath).Call("org.gnome.Shell.SearchProvider2.AddDocument", 0, path, title, authors, keywords).Store()
	if err != nil {
		return err
	}

	return nil
}
