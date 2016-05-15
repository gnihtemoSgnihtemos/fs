package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/martinp/ftpsc/crawler"
	"github.com/martinp/ftpsc/ftp"
)

func main() {
	var opts struct {
		Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.ftpscrc"`
		Test   bool   `short:"t" long:"test" description:"Test and print config"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	var name string
	if opts.Config == "~/.ftpscrc" {
		home := os.Getenv("HOME")
		name = filepath.Join(home, ".ftpscrc")
	} else {
		name = opts.Config
	}

	cfg, err := crawler.ReadConfig(name)
	if err != nil {
		log.Fatal(err)
	}

	for _, site := range cfg.Sites {
		logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", site.Name), log.LstdFlags)
		ftpClient, err := ftp.DialTimeout("tcp", site.Address, time.Second*site.ConnectTimeout)
		defer ftpClient.Quit()
		if err != nil {
			logger.Printf("connection timed out after %d seconds: %s", site.ConnectTimeout, err)
			continue
		}
		if site.TLS {
			if err := ftpClient.LoginWithTLS(&tls.Config{InsecureSkipVerify: true}, site.Username, site.Password); err != nil {
				logger.Printf("login with TLS failed: %s", err)
				continue
			}
		} else {
			if err := ftpClient.Login(site.Username, site.Password); err != nil {
				logger.Printf("login failed: %s", err)
				continue
			}
		}

		c := crawler.New(ftpClient, site, logger)
		dirs, err := c.WalkDirs(site.Root)
		if err != nil {
			logger.Printf("failed walking directories: %s", err)
			continue
		}
		for _, d := range dirs {
			fmt.Printf("%+v\n", d)
		}
	}
}
