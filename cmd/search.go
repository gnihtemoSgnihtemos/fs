package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/martinp/fs/database"
	"github.com/olekukonko/tablewriter"
)

type Search struct {
	opts
	Site   string   `short:"s" long:"site" description:"Search a specific site" value-name:"NAME"`
	Limit  int      `short:"c" long:"max-count" description:"Maximum number of results to show"`
	Format string   `short:"F" long:"format" description:"Format to use when printing results" choice:"table" choice:"simple" choice:"path" default:"table"`
	Order  []string `short:"o" long:"order" description:"Columns to sort results by" value-name:"COLUMN" default:"site:asc" default:"dir.path:asc"`
}

func writeTable(w io.Writer, dirs []database.Dir) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Site", "Path", "Date"})
	for _, d := range dirs {
		date := time.Unix(d.Modified, 0).UTC().Format("2006-01-02 15:04:05 MST")
		row := []string{d.Site, d.Path, date}
		table.Append(row)
	}
	table.Render()
	return nil
}

func writeSimple(w io.Writer, dirs []database.Dir) error {
	tab := tabwriter.NewWriter(w, 0, 8, 0, '\t', 0)
	fmt.Fprintln(tab, "SITE\tPATH\tDATE")
	for _, d := range dirs {
		fmt.Fprintf(tab, "%s\t%s\t%d\n", d.Site, d.Path, d.Modified)
	}
	return tab.Flush()
}

func writePath(w io.Writer, dirs []database.Dir) error {
	for _, d := range dirs {
		fmt.Fprintln(w, d.Path)
	}
	return nil
}

func (c *Search) Execute(args []string) error {
	cfg := mustReadConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	order, err := database.OrderByClauses(c.Order)
	if err != nil {
		return err
	}
	keywords := strings.Join(args, " ")
	dirs, err := db.SelectDirs(keywords, c.Site, order, c.Limit)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no results found")
	}
	switch c.Format {
	case "simple":
		return writeSimple(os.Stdout, dirs)
	case "path":
		return writePath(os.Stdout, dirs)
	}
	return writeTable(os.Stdout, dirs)
}
