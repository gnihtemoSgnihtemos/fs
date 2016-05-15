package database

import (
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
)

const schema = `
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
  CONSTRAINT path_unique UNIQUE(site_id, path),
  FOREIGN KEY(site_id) REFERENCES site(id) ON DELETE CASCADE
);
`

type Site struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Dir struct {
	Site string `db:"site"`
	Path string `db:"path"`
	Name string `db:"name"`
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
	for _, s := range sites {
		if _, err := tx.Exec("DELETE FROM site WHERE id=$1", s.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
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
	for _, d := range dirs {
		if _, err := tx.Exec("INSERT INTO dir (site_id, path, name) VALUES ($1, $2, $3)", site.ID, d.Path, d.Name); err != nil {
			return err
		}
	}
	return tx.Commit()
}
