package cmd

import (
	"log"
	"os"

	"github.com/martinp/fs/crawler"
	"github.com/martinp/fs/database"
)

type Update struct {
	opts
	Site string `short:"s" long:"site" description:"Update a single site" value-name:"NAME"`
}

func (c *Update) Execute(args []string) error {
	if len(args) != 0 {
		return errUnexpectedArgs
	}
	cfg := mustReadConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	logger := log.New(os.Stderr, "", log.LstdFlags)
	sem := make(chan bool, cfg.Concurrency)
	for _, site := range cfg.Sites {
		if c.Site != "" && c.Site != site.Name {
			continue
		}
		sem <- true
		go func(site crawler.Site) {
			defer func() { <-sem }()
			c := crawler.New(site, db, logger)
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
