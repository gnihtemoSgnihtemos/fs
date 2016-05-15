package cmd

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/martinp/fs/database"
	"github.com/olekukonko/tablewriter"
)

type Search struct{ opts }

func (c *Search) WriteTable(dirs []database.Dir, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Site", "Path", "Date"})
	for _, d := range dirs {
		date := time.Unix(d.Modified, 0).Format("2006-01-02 15:04:05 MST")
		row := []string{d.Site, d.Path, date}
		table.Append(row)
	}
	table.Render()
}

func (c *Search) Execute(args []string) error {
	cfg := readConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	dirs, err := db.FindDirs(strings.Join(args, " "))
	if err != nil {
		return err
	}
	c.WriteTable(dirs, os.Stdout)
	return nil
}
