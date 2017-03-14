package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/mpolden/fs/crawler"
)

var errUnexpectedArgs = errors.New("command does not accept any arguments")

type opts struct {
	Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.fsrc"`
}

func mustReadConfig(name string) crawler.Config {
	if name == "~/.fsrc" {
		home := os.Getenv("HOME")
		name = filepath.Join(home, ".fsrc")
	}
	cfg, err := crawler.ReadConfig(name)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
