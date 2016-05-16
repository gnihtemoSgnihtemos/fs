package crawler

import (
	"crypto/tls"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/martinp/fs/database"
	"github.com/martinp/fs/ftp"
)

type dirLister interface {
	List(path string) ([]ftp.File, error)
	FilterFiles([]ftp.File) []ftp.File
}

type Crawler struct {
	site      Site
	logger    *log.Logger
	ftpClient *ftp.Client
	dbClient  *database.Client
}

func New(site Site, dbClient *database.Client, logger *log.Logger) *Crawler {
	return &Crawler{
		dbClient: dbClient,
		site:     site,
		logger:   logger,
	}
}

func (c *Crawler) Connect() error {
	ftpClient, err := ftp.DialTimeout("tcp", c.site.Address, time.Second*c.site.ConnectTimeout)
	if err != nil {
		return err
	}
	if c.site.TLS {
		if err := ftpClient.LoginWithTLS(&tls.Config{InsecureSkipVerify: true}, c.site.Username, c.site.Password); err != nil {
			return err
		}
	} else {
		if err := ftpClient.Login(c.site.Username, c.site.Password); err != nil {
			return err
		}
	}
	c.ftpClient = ftpClient
	c.Logf("Connected to %s (TLS=%t)", c.site.Address, c.site.TLS)
	return nil
}

func (c *Crawler) Close() error {
	return c.ftpClient.Quit()
}

func (c *Crawler) Logf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("[%s] ", c.site.Name)
	c.logger.Printf(prefix+format, v...)
}

func (c *Crawler) List(path string) ([]ftp.File, error) {
	message, err := c.ftpClient.Stat(path)
	if err != nil {
		c.Logf("Listing directory %s failed: %s", path, err)
		return nil, nil
	}
	return ftp.ParseFiles(path, strings.NewReader(message))
}

func (c *Crawler) FilterFiles(files []ftp.File) []ftp.File {
	return filterFiles(files, c.site.Ignore, c.site.IgnoreSymlinks)
}

func (c *Crawler) WalkShallow(path string) ([]ftp.File, error) {
	return walkShallow(c, path, -1)
}

func (c *Crawler) Run() error {
	c.Logf("Walking %s", c.site.Root)
	files, err := c.WalkShallow(c.site.Root)
	if err != nil {
		return err
	}
	c.Logf("Mapping %d files to directories", len(files))
	dirs := makeDirs(files)
	c.Logf("Removing existing directories from database")
	if err := c.dbClient.DeleteDirs(c.site.Name); err != nil {
		return err
	}
	c.Logf("Inserting %d directories into database", len(dirs))
	if err := c.dbClient.Insert(c.site.Name, dirs); err != nil {
		return err
	}
	c.Logf("Saved %d directories", len(dirs))
	return nil
}

func filterFiles(files []ftp.File, ignore []string, ignoreSymlink bool) []ftp.File {
	keep := []ftp.File{}
Loop:
	for _, f := range files {
		if ignoreSymlink && f.IsSymlink() {
			continue
		}
		for _, name := range ignore {
			if f.Name == name {
				continue Loop
			}
		}
		keep = append(keep, f)
	}
	return keep
}

func makeDirs(files []ftp.File) []database.Dir {
	keep := []database.Dir{}
	for _, f := range files {
		if f.IsCurrentOrParent() {
			continue
		}
		d := database.Dir{
			Name:     f.Name,
			Path:     f.Path,
			Modified: f.Modified.Unix(),
		}
		keep = append(keep, d)
	}
	return keep
}

func walkShallow(lister dirLister, path string, maxdepth int) ([]ftp.File, error) {
	files, err := lister.List(path)
	files = lister.FilterFiles(files)
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
		peek = lister.FilterFiles(peek)
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
