package crawler

import (
	"crypto/tls"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/martinp/ftpsc/database"
	"github.com/martinp/ftpsc/ftp"
)

type dirLister interface {
	List(path string) ([]ftp.File, error)
}

type Crawler struct {
	site      Site
	log       *log.Logger
	ftpClient *ftp.Client
	dbClient  *database.Client
}

func newCrawler(site Site, ftpClient *ftp.Client, dbClient *database.Client, logger *log.Logger) *Crawler {
	return &Crawler{
		ftpClient: ftpClient,
		dbClient:  dbClient,
		site:      site,
		log:       logger,
	}
}

func Connect(site Site, dbClient *database.Client, logger *log.Logger) (*Crawler, error) {
	ftpClient, err := ftp.DialTimeout("tcp", site.Address, time.Second*site.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	if site.TLS {
		if err := ftpClient.LoginWithTLS(&tls.Config{InsecureSkipVerify: true}, site.Username, site.Password); err != nil {
			return nil, err
		}
	} else {
		if err := ftpClient.Login(site.Username, site.Password); err != nil {
			return nil, err
		}
	}
	logger.Print("Connected")
	return newCrawler(site, ftpClient, dbClient, logger), nil
}

func (c *Crawler) List(path string) ([]ftp.File, error) {
	message, err := c.ftpClient.Stat("-al " + path)
	if err != nil {
		c.log.Printf("listing directory %s failed: %s", path, err)
		return nil, nil
	}
	return ftp.ParseFiles(path, strings.NewReader(message))
}

func (c *Crawler) WalkShallow(path string) ([]ftp.File, error) {
	return walkShallow(c, path, -1)
}

func (c *Crawler) Run() error {
	files, err := c.WalkShallow(c.site.Root)
	if err != nil {
		return err
	}
	keep := []database.Dir{}
	for _, f := range files {
		if f.IsCurrentOrParent() {
			continue
		}
		d := database.Dir{Name: f.Name, Path: f.Path}
		keep = append(keep, d)
	}
	if err := c.dbClient.DeleteDirs(c.site.Name); err != nil {
		return err
	}
	if err := c.dbClient.Insert(c.site.Name, keep); err != nil {
		return err
	}
	c.log.Printf("saved %d entries", len(keep))
	return nil
}

func walkShallow(lister dirLister, path string, maxdepth int) ([]ftp.File, error) {
	files, err := lister.List(path)
	if err != nil {
		return nil, err
	}
Loop:
	for _, f := range files {
		if f.IsCurrentOrParent() {
			continue
		}
		if !f.Mode.IsDir() {
			continue
		}
		subpath := filepath.Join(path, f.Name)
		depth := strings.Count(subpath, "/")
		if maxdepth > 0 && depth > maxdepth {
			continue
		}
		// Peek at sub-directory to determine max depth
		peek, err := lister.List(subpath)
		if err != nil {
			return nil, err
		}
		for _, f := range peek {
			if !f.Mode.IsDir() {
				maxdepth = depth - 1
				continue Loop
			}
		}
		fs, err := walkShallow(lister, subpath, maxdepth)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}
