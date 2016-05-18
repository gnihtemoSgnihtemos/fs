package cmd

import (
	"log"

	"github.com/martinp/fs/database"
)

type GC struct {
	opts
	Dryrun bool `short:"n" long:"dry-run" description:"Only show what would be deleted"`
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
	remove := []database.Site{}
	for _, s1 := range sites {
		found := false
		for _, s2 := range cfg.Sites {
			if s1.Name == s2.Name {
				found = true
				break
			}
		}
		if !found {
			remove = append(remove, s1)
		}
	}
	if c.Dryrun {
		for _, s := range remove {
			log.Printf("Would remove %s\n", s.Name)
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
