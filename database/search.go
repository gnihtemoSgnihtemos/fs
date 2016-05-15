package database

import (
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
)

func (c *Client) WriteTable(dirs []Dir, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Site", "Path", "Date"})
	for _, d := range dirs {
		date := time.Unix(d.Modified, 0).Format("2006-01-02 15:04:05 MST")
		row := []string{d.Site, d.Path, date}
		table.Append(row)
	}
	table.Render()
}
