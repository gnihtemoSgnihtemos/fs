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

CREATE INDEX IF NOT EXISTS dir_site_id_idx ON dir (site_id);

-- FTS index table
CREATE VIRTUAL TABLE IF NOT EXISTS dir_fts USING fts4(
  id INTEGER PRIMARY KEY,
  site_id INTEGER,
  path TEXT
);

-- Triggers to keep FTS table up to date
CREATE TRIGGER IF NOT EXISTS dir_bd BEFORE DELETE ON dir BEGIN
  DELETE FROM dir_fts WHERE docid=old.rowid;
END;
CREATE TRIGGER IF NOT EXISTS dir_ai AFTER INSERT ON dir BEGIN
  INSERT INTO dir_fts(id, site_id, path) VALUES (new.id, new.site_id, new.path);
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
	// Ensure foreign keys are enabled (defaults to off)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

func (c *Client) insertSite(siteName string) error {
	_, err := c.db.Exec("INSERT INTO site (name) VALUES ($1)", siteName)
	return err

}

func (c *Client) getOrInsertSite(siteName string) (Site, error) {
	var site Site
	err := c.db.Get(&site, "SELECT * FROM site WHERE name = $1", siteName)
	if err == sql.ErrNoRows {
		if err := c.insertSite(siteName); err != nil {
			return Site{}, err
		}
		return c.getOrInsertSite(siteName)
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

func (c *Client) DeleteSites(sites []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, s := range sites {
		if _, err := tx.Exec("DELETE FROM site WHERE name = $1", s); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) DeleteSite(site string) error {
	return c.DeleteSites([]string{site})
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
	site, err := c.getOrInsertSite(siteName)
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

func (c *Client) SelectDirs(keywords string) ([]Dir, error) {
	query := `SELECT site.name AS site, dir_fts.path, dir.name, dir.modified FROM dir_fts
                  INNER JOIN dir ON dir_fts.id = dir.id
                  INNER JOIN site ON dir_fts.site_id = site.id
                  WHERE dir_fts.path MATCH $1`
	var dirs []Dir
	if err := c.db.Select(&dirs, query, keywords); err != nil {
		return nil, err
	}
	return dirs, nil
}

func (c *Client) SelectDirsSite(site string, keywords string) ([]Dir, error) {
	query := `SELECT site.name AS site, dir_fts.path, dir.name, dir.modified FROM dir_fts
                  INNER JOIN dir ON dir_fts.id = dir.id
                  INNER JOIN site ON dir_fts.site_id = site.id
                  WHERE dir_fts.path MATCH $1 AND site.name = $2`
	var dirs []Dir
	if err := c.db.Select(&dirs, query, keywords, site); err != nil {
		return nil, err
	}
	return dirs, nil
}
