package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mpolden/fs/cmd"
)

func main() {
	p := flags.NewParser(nil, flags.Default)

	logger := log.New(os.Stderr, "fs: ", 0)
	var update cmd.Update
	if _, err := p.AddCommand("update", "Update database",
		"Crawls sites and updates the database.", &update); err != nil {
		log.Fatal(err)
	}
	update.Logger = logger
	var gc cmd.GC
	if _, err := p.AddCommand("gc", "Clean database",
		"Remove entries for sites that do not exist in config", &gc); err != nil {
		log.Fatal(err)
	}
	gc.Logger = logger
	var search cmd.Search
	if _, err := p.AddCommand("search", "Search database",
		"Search database", &search); err != nil {
		log.Fatal(err)
	}
	var test cmd.Test
	if _, err := p.AddCommand("test", "Test configuration",
		"Test and print configuration", &test); err != nil {
		log.Fatal(err)
	}
	if _, err := p.Parse(); err != nil {
		os.Exit(1)
	}
}
