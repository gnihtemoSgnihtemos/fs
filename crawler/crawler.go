package crawler

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/martinp/ftpsc/database"
	"github.com/martinp/ftpsc/ftp"
)

type dirLister interface {
	List(path string) ([]ftp.File, error)
}

type Client struct {
	site      Site
	log       *log.Logger
	ftpClient *ftp.Client
	dbClient  *database.Client
}

func New(ftpClient *ftp.Client, dbClient *database.Client, site Site, logger *log.Logger) *Client {
	return &Client{
		ftpClient: ftpClient,
		dbClient:  dbClient,
		site:      site,
		log:       logger,
	}
}

func (c *Client) List(path string) ([]ftp.File, error) {
	message, err := c.ftpClient.Stat("-al " + path)
	if err != nil {
		c.log.Printf("listing directory %s failed: %s", path, err)
		return nil, nil
	}
	return ftp.ParseFiles(path, strings.NewReader(message))
}

func (c *Client) WalkDirs(path string) ([]ftp.File, error) {
	return walkDirs(c, path, -1)
}

func (c *Client) Run() error {
	files, err := c.WalkDirs(c.site.Root)
	if err != nil {
		return err
	}
	keep := []ftp.File{}
	for _, f := range files {
		if f.IsCurrentOrParent() {
			continue
		}
		keep = append(keep, f)
	}
	if err := c.dbClient.Insert(c.site.Name, keep); err != nil {
		return err
	}
	c.log.Printf("saved %d entries", len(keep))
	return nil
}

func walkDirs(lister dirLister, path string, maxdepth int) ([]ftp.File, error) {
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
		fs, err := walkDirs(lister, subpath, maxdepth)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}
