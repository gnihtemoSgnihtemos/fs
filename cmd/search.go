package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mpolden/fs/database"
	"github.com/olekukonko/tablewriter"
)

type Search struct {
	opts
	Site   string   `short:"s" long:"site" description:"Search a specific site" value-name:"NAME"`
	Limit  int      `short:"c" long:"max-count" description:"Maximum number of results to show"`
	Format string   `short:"F" long:"format" description:"Format to use when printing results. (default: table, when piping: path)" choice:"table" choice:"simple" choice:"path"`
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

func writeResults(f *os.File, format string, dirs []database.Dir) error {
	if format == "" {
		stat, err := f.Stat()
		if err != nil {
			return err
		}
		// Default to path format if ouput is being piped
		if stat.Mode()&os.ModeCharDevice == 0 {
			format = "path"
		}
	}
	switch format {
	case "simple":
		return writeSimple(f, dirs)
	case "path":
		return writePath(f, dirs)
	}
	return writeTable(f, dirs)
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
	return writeResults(os.Stdout, c.Format, dirs)
}
