package database

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
-- Tables for sites and dirs
CREATE TABLE IF NOT EXISTS site (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  CONSTRAINT name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS dir (
  id INTEGER PRIMARY KEY,
  site_id INTEGER NOT NULL,
  path TEXT NOT NULL,
  modified INTEGER NOT NULL,
  CONSTRAINT path_unique UNIQUE(site_id, path),
  FOREIGN KEY(site_id) REFERENCES site(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS dir_site_id_idx ON dir (site_id);

-- FTS index table
CREATE VIRTUAL TABLE IF NOT EXISTS dir_fts USING fts4(
  id INTEGER PRIMARY KEY,
  site_id INTEGER NOT NULL,
  path TEXT NOT NULL
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

func (c *Client) SelectSites() ([]Site, error) {
	var sites []Site
	if err := c.db.Select(&sites, "SELECT * FROM site ORDER BY name ASC"); err != nil {
		return nil, err
	}
	return sites, nil
}

func (c *Client) DeleteSites(names []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, n := range names {
		if _, err := tx.Exec("DELETE FROM site WHERE name = $1", n); err != nil {
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

func (c *Client) Optimize() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.db.Exec("INSERT INTO dir_fts (dir_fts) VALUES ('optimize')")
	return err
}

func (c *Client) Insert(siteName string, dirs []Dir) error {
	// Ensure writes to SQLite db are serialized
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM site WHERE name = $1", siteName); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO site (name) VALUES ($1)", siteName); err != nil {
		return err
	}
	siteID := 0
	if err := tx.Get(&siteID, "SELECT id FROM site WHERE name = $1", siteName); err != nil {
		return err
	}
	for _, d := range dirs {
		if _, err := tx.Exec("INSERT INTO dir (site_id, path, modified) VALUES ($1, $2, $3)", siteID, d.Path, d.Modified); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func OrderByClause(s string) (string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 0 {
		return "", fmt.Errorf("could not parse order: %s", s)
	}
	field := parts[0]
	order := ""
	if len(parts) == 2 {
		order = parts[1]
	}
	orderUpper := strings.ToUpper(order)
	switch orderUpper {
	case "":
	case "ASC":
	case "DESC":
		break
	default:
		return "", fmt.Errorf("invalid order: %s", order)
	}
	if order == "" {
		return field, nil
	}
	return field + " " + orderUpper, nil
}

func OrderByClauses(ss []string) (string, error) {
	var orderByClauses []string
	for _, s := range ss {
		c, err := OrderByClause(s)
		if err != nil {
			return "", err
		}
		orderByClauses = append(orderByClauses, c)
	}
	return strings.Join(orderByClauses, ", "), nil
}

func selectDirsQuery(keywords, site, order string, limit int) (string, []interface{}) {
	query := `SELECT site.name AS site, dir_fts.path, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1`
	args := []interface{}{keywords}
	if site != "" {
		query += " AND site.name = $2"
		args = append(args, site)
	}
	if order != "" {
		query += " ORDER BY " + order
	}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query, args
}

func (c *Client) SelectDirs(keywords, site, order string, limit int) ([]Dir, error) {
	query, args := selectDirsQuery(keywords, site, order, limit)
	var dirs []Dir
	if err := c.db.Select(&dirs, query, args...); err != nil {
		return nil, err
	}
	return dirs, nil
}
