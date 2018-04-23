package crawler

import (
	"crypto/tls"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mpolden/fs/database"
	"github.com/mpolden/fs/ftp"
)

type dirLister interface {
	list(path string) ([]ftp.File, error)
	filterFiles([]ftp.File) []ftp.File
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
	var ftpClient *ftp.Client
	var err error
	if c.site.proxyURL != nil {
		ftpClient, err = ftp.DialWithProxy("tcp", c.site.Address, c.site.proxyURL, c.site.connectTimeout)
	} else {
		ftpClient, err = ftp.DialTimeout("tcp", c.site.Address, c.site.connectTimeout)
	}
	if err != nil {
		return err
	}
	ftpClient.ReadTimeout = c.site.readTimeout
	if c.site.TLS {
		err = ftpClient.LoginWithTLS(&tls.Config{InsecureSkipVerify: true}, c.site.Username, c.site.Password)
	} else {
		err = ftpClient.Login(c.site.Username, c.site.Password)
	}
	if err != nil {
		return err
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

func (c *Crawler) list(path string) ([]ftp.File, error) {
	message, err := c.ftpClient.Stat(path)
	if err != nil {
		c.Logf("Listing directory %s failed: %s", path, err)
		return nil, nil
	}
	return ftp.ParseFiles(path, strings.NewReader(message))
}

func (c *Crawler) filterFiles(files []ftp.File) []ftp.File {
	return filterFiles(files, c.site.Ignore, c.site.IgnoreSymlinks)
}

func (c *Crawler) walk(path string) ([]ftp.File, error) {
	return walk(c, path, -1)
}

func (c *Crawler) Run() error {
	c.Logf("Walking %s", c.site.Root)
	files, err := c.walk(c.site.Root)
	if err != nil {
		return err
	}
	c.Logf("Removing existing directories from database")
	if err := c.dbClient.DeleteSite(c.site.Name); err != nil {
		return err
	}
	dirs := toDirs(files)
	c.Logf("Inserting %d directories into database", len(dirs))
	if err := c.dbClient.Insert(c.site.Name, dirs); err != nil {
		return err
	}
	c.Logf("Saved %d directories", len(dirs))
	return nil
}

func filterFiles(files []ftp.File, excludedFiles []string, ignoreSymlinks bool) []ftp.File {
	keep := []ftp.File{}
Loop:
	for _, f := range files {
		if ignoreSymlinks && f.IsSymlink() {
			continue
		}
		for _, name := range excludedFiles {
			if f.Name == name {
				continue Loop
			}
		}
		keep = append(keep, f)
	}
	return keep
}

func toDirs(files []ftp.File) []database.Dir {
	keep := []database.Dir{}
	for _, f := range files {
		d := database.Dir{
			Path:     f.Path,
			Modified: f.Modified.Unix(),
		}
		keep = append(keep, d)
	}
	return keep
}

func list(lister dirLister, path string) ([]ftp.File, error) {
	files, err := lister.list(path)
	if err != nil {
		return nil, err
	}
	files = lister.filterFiles(files)
	sort.Slice(files, func(i, j int) bool {
		// Sort file names starting with underscore first
		if strings.Index(files[i].Name, "_") == 0 && strings.Index(files[j].Name, "_") != 0 {
			return true
		}
		return files[i].Name < files[j].Name
	})
	return files, nil
}

func containsOnlyDir(files []ftp.File) bool {
	for _, f := range files {
		if !f.Mode.IsDir() {
			return false
		}
	}
	return true
}

func walk(lister dirLister, path string, maxdepth int) ([]ftp.File, error) {
	files, err := list(lister, path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.Mode.IsDir() {
			continue
		}
		subpath := filepath.Join(path, f.Name)
		depth := strings.Count(subpath, "/")
		if maxdepth > 0 && depth > maxdepth {
			continue
		}
		// Peek at sub-directory to determine max depth
		children, err := list(lister, subpath)
		if err != nil {
			return nil, err
		}
		if !containsOnlyDir(children) {
			maxdepth = depth - 1
			continue
		}
		fs, err := walk(lister, subpath, maxdepth)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}
