package cmd

import (
	"log"

	"github.com/martinp/fs/crawler"
	"github.com/martinp/fs/database"
)

type GC struct {
	opts
	Dryrun bool `short:"n" long:"dry-run" description:"Only show what would be deleted"`
}

func difference(sites []database.Site, configSites []crawler.Site) []string {
	var diff []string
	for _, s1 := range sites {
		found := false
		for _, s2 := range configSites {
			if s1.Name == s2.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1.Name)
		}
	}
	return diff
}

func (c *GC) Execute(args []string) error {
	cfg := mustReadConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	sites, err := db.GetSites()
	if err != nil {
		return err
	}
	remove := difference(sites, cfg.Sites)
	if c.Dryrun {
		for _, s := range remove {
			log.Printf("Would remove %s\n", s)
		}
		return nil
	}
	log.Printf("Removing %d sites", len(remove))
	if err := db.DeleteSites(remove); err != nil {
		return err
	}
	log.Print("Running vacuum")
	return db.Vacuum()
}
