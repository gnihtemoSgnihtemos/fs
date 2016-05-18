package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/martinp/fs/database"
	"github.com/olekukonko/tablewriter"
)

type Search struct {
	opts
	Site  string `short:"s" long:"site" description:"Search a specific site" value-name:"NAME"`
	Limit int    `short:"c" long:"max-count" description:"Maximum number of results to show"`
}

func (c *Search) writeTable(dirs []database.Dir, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Site", "Path", "Date"})
	for _, d := range dirs {
		date := time.Unix(d.Modified, 0).UTC().Format("2006-01-02 15:04:05 MST")
		row := []string{d.Site, d.Path, date}
		table.Append(row)
	}
	table.Render()
}

func (c *Search) Execute(args []string) error {
	cfg := mustReadConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	keywords := strings.Join(args, " ")
	dirs, err := db.SelectDirs(keywords, c.Site, c.Limit)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no results found")
	}
	c.writeTable(dirs, os.Stdout)
	return nil
}
