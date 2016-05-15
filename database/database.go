package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"

	"github.com/martinp/ftpsc/ftp"
)

const schema = `
CREATE TABLE IF NOT EXISTS site (
  id INTEGER PRIMARY KEY,
  name TEXT,
  CONSTRAINT name_unique UNIQUE (name)
);
CREATE TABLE IF NOT EXISTS entry (
  id INTEGER PRIMARY KEY,
  site_id INTEGER,
  path TEXT,
  name TEXT,
  CONSTRAINT path_unique UNIQUE(site_id, path),
  FOREIGN KEY(site_id) REFERENCES site(id)
);
`

type Site struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Entry struct {
	Site string `db:"site"`
	Path string `db:"path"`
	Name string `db:"name"`
}

type Client struct {
	db *sqlx.DB
}

func New(filename string) (*Client, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
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

func (c *Client) addSite(siteName string) error {
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
		if err := c.addSite(siteName); err != nil {
			return Site{}, err
		}
		return c.getSite(siteName)
	} else if err != nil {
		return Site{}, err
	}
	return site, nil
}

func (c *Client) Add(siteName string, files []ftp.File) error {
	site, err := c.getSite(siteName)
	if err != nil {
		return err
	}
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	for _, f := range files {
		if _, err := tx.Exec("INSERT INTO entry (site_id, path, name) VALUES ($1, $2, $3)", site.ID, f.Path, f.Name); err != nil {
			return err
		}
	}
	return tx.Commit()
}
