package database

import (
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
)

const schema = `
-- Tables for sites and dirs
CREATE TABLE IF NOT EXISTS site (
  id INTEGER PRIMARY KEY,
  name TEXT,
  CONSTRAINT name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS dir (
  id INTEGER PRIMARY KEY,
  site_id INTEGER,
  path TEXT,
  name TEXT,
  modified INTEGER,
  CONSTRAINT path_unique UNIQUE(site_id, path),
  FOREIGN KEY(site_id) REFERENCES site(id) ON DELETE CASCADE
);

-- FTS index table
CREATE VIRTUAL TABLE IF NOT EXISTS dir_fts USING fts4(
  id INTEGER PRIMARY KEY,
  site_id INTEGER,
  path TEXT,
  name TEXT,
  modified INTEGER
);

-- Triggers to keep FTS table up to date
CREATE TRIGGER IF NOT EXISTS dir_bd BEFORE DELETE ON dir BEGIN
  DELETE FROM dir_fts WHERE id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS dir_ai AFTER INSERT ON dir BEGIN
  INSERT INTO dir_fts(id, site_id, path, name, modified) VALUES (new.id, new.site_id, new.path, new.name, new.modified);
END;
`

type Site struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Dir struct {
	Site     string `db:"site"`
	Path     string `db:"path"`
	Name     string `db:"name"`
	Modified int64  `db:"modified"`
}

type Client struct {
	db *sqlx.DB
	mu sync.Mutex
}

func New(filename string) (*Client, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	// Enable foregin keys support
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) insertSite(siteName string) error {
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("INSERT INTO site (name) VALUES ($1)", siteName); err != nil {
		return err
	}
	return tx.Commit()
}

func (c *Client) getSite(siteName string) (Site, error) {
	var site Site
	err := c.db.Get(&site, "SELECT * FROM site WHERE name=$1", siteName)
	if err == sql.ErrNoRows {
		if err := c.insertSite(siteName); err != nil {
			return Site{}, err
		}
		return c.getSite(siteName)
	} else if err != nil {
		return Site{}, err
	}
	return site, nil
}

func (c *Client) GetSites() ([]Site, error) {
	var sites []Site
	if err := c.db.Select(&sites, "SELECT * FROM site ORDER BY name ASC"); err != nil {
		return nil, err
	}
	return sites, nil
}

func (c *Client) DeleteSites(sites []Site) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, s := range sites {
		if _, err := tx.Exec("DELETE FROM site WHERE id=$1", s.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) DeleteDirs(site string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.db.Exec("DELETE FROM dir WHERE site_id IN (SELECT id FROM site WHERE name = $1)", site)
	return err
}

func (c *Client) Vacuum() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.db.Exec("VACUUM")
	return err
}

func (c *Client) Insert(siteName string, dirs []Dir) error {
	// Ensure writes to SQLite db are serialized
	c.mu.Lock()
	defer c.mu.Unlock()
	site, err := c.getSite(siteName)
	if err != nil {
		return err
	}
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, d := range dirs {
		if _, err := tx.Exec("INSERT INTO dir (site_id, path, name, modified) VALUES ($1, $2, $3, $4)", site.ID, d.Path, d.Name, d.Modified); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) FindDirs(keyword string) ([]Dir, error) {
	var dirs []Dir
	if err := c.db.Select(&dirs, "SELECT site.name AS site, path, dir_fts.name, modified FROM dir_fts INNER JOIN site ON site_id = site.id WHERE path MATCH $1", keyword); err != nil {
		return nil, err
	}
	return dirs, nil
}

func (c *Client) FindDirsBySite(keyword string, site string) ([]Dir, error) {
	var dirs []Dir
	if err := c.db.Select(&dirs, "SELECT site.name AS site, path, dir_fts.name, modified FROM dir_fts INNER JOIN site ON site_id = site.id WHERE path MATCH $1 AND site.name = $2", keyword, site); err != nil {
		return nil, err
	}
	return dirs, nil
}
