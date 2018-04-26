package cmd

import (
	"log"
	"os"

	"github.com/mpolden/fs/crawler"
	"github.com/mpolden/fs/database"
)

type Update struct {
	opts
	Dryrun bool     `short:"n" long:"dry-run" description:"Only show what would be crawled"`
	Sites  []string `short:"s" long:"site" description:"Update a single site" value-name:"NAME"`
}

func (u *Update) updateSite(name string) bool {
	for _, site := range u.Sites {
		if site == name {
			return true
		}
	}
	return len(u.Sites) == 0
}

func (u *Update) Execute(args []string) error {
	if len(args) != 0 {
		return errUnexpectedArgs
	}
	cfg := mustReadConfig(u.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	logger := log.New(os.Stderr, "fs: ", 0)
	sem := make(chan bool, cfg.Concurrency)
	for _, site := range cfg.Sites {
		if !u.updateSite(site.Name) || site.Skip {
			continue
		}
		sem <- true
		go func(site crawler.Site) {
			defer func() { <-sem }()
			c := crawler.New(site, db, logger)
			if u.Dryrun {
				c.Logf("Updating (dry run)")
				return
			}
			if err := c.Connect(); err != nil {
				c.Logf("Failed to connect: %s", err)
				return
			}
			defer c.Close()
			if err := c.Run(); err != nil {
				c.Logf("Failed crawling: %s", err)
				return
			}
		}(site)
	}
	// Wait for remaining goroutines to finish
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return nil
}
