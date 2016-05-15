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
	Site string `short:"s" long:"site" description:"Search a specific site" value-name:"NAME"`
}

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
	keywords := strings.Join(args, " ")
	var dirs []database.Dir
	if c.Site == "" {
		d, err := db.FindDirs(keywords)
		if err != nil {
			return err
		}
		dirs = d
	} else {
		d, err := db.FindDirsBySite(keywords, c.Site)
		if err != nil {
			return err
		}
		dirs = d
	}
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no results found")
	}
	c.WriteTable(dirs, os.Stdout)
	return nil
}
