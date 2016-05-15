package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/martinp/ftpsc/crawler"
)

type opts struct {
	Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.ftpscrc"`
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
