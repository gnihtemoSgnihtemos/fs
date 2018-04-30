package cmd

import (
	"log"

	"github.com/mpolden/fs/database"
)

type GC struct {
	opts
	Logger  *log.Logger
	Dryrun  bool     `short:"n" long:"dry-run" description:"Only show what would be deleted"`
	Exclude []string `short:"e" long:"exclude" description:"Exclude sites" value-name:"SITES"`
}

func difference(ss1 []string, ss2 []string) []string {
	var diff []string
	for _, s1 := range ss1 {
		found := false
		for _, s2 := range ss2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}
	return diff
}

func (c *GC) Execute(args []string) error {
	if len(args) != 0 {
		return errUnexpectedArgs
	}
	cfg := mustReadConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	sites, err := db.SelectSites()
	if err != nil {
		return err
	}
	var dbSites []string
	for _, s := range sites {
		dbSites = append(dbSites, s.Name)
	}
	var cfgSites []string
	for _, s := range cfg.Sites {
		cfgSites = append(cfgSites, s.Name)
	}
	remove := difference(difference(dbSites, cfgSites), c.Exclude)
	if c.Dryrun {
		for _, s := range remove {
			log.Printf("Would remove %s\n", s)
		}
		return nil
	}
	c.Logger.Printf("Removing %d sites", len(remove))
	if err := db.DeleteSites(remove); err != nil {
		return err
	}
	c.Logger.Print("Optimizing FTS table")
	if err := db.Optimize(); err != nil {
		return err
	}
	c.Logger.Print("Rebuilding database")
	return db.Vacuum()
}
