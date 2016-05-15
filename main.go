package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/martinp/ftpsc/crawler"
	"github.com/martinp/ftpsc/database"
)

type opts struct {
	Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.ftpscrc"`
}

type updateCmd struct {
	opts
	Site string `short:"s" long:"site" description:"Update a single site" value-name:"NAME"`
}

type gcCmd struct {
	opts
	Dryrun bool `short:"n" long:"dry-run" description:"Only show what would be deleted"`
}

func (c *updateCmd) Execute(args []string) error {
	cfg := readConfig(c.Config)
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	sem := make(chan bool, cfg.Concurrency)
	for _, site := range cfg.Sites {
		sem <- true
		go func(site crawler.Site) {
			defer func() { <-sem }()
			logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", site.Name), log.LstdFlags)
			c, err := crawler.Connect(site, db, logger)
			if err != nil {
				logger.Printf("Failed to connect: %s", err)
				return
			}
			if err := c.Run(); err != nil {
				logger.Printf("Failed crawling: %s", err)
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

func (c *gcCmd) Execute(args []string) error {
	cfg := readConfig(c.Config)
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
			fmt.Printf("Deleting %s\n", s.Name)
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

func readConfig(name string) crawler.Config {
	if name == "~/.ftpscrc" {
		home := os.Getenv("HOME")
		name = filepath.Join(home, ".ftpscrc")
	}
	cfg, err := crawler.ReadConfig(name)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func main() {
	p := flags.NewParser(nil, flags.Default)

	var update updateCmd
	if _, err := p.AddCommand("update", "Update database",
		"Crawls sites and updates the database.", &update); err != nil {
		log.Fatal(err)
	}
	var gc gcCmd
	if _, err := p.AddCommand("gc", "Clean database",
		"Remove entries for sites that do not exist in config", &gc); err != nil {
		log.Fatal(err)
	}
	if _, err := p.Parse(); err != nil {
		log.Fatal(err)
	}
}
