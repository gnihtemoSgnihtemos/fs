package database

import (
	"reflect"
	"testing"
)

func testClient() *Client {
	c, err := New(":memory:")
	if err != nil {
		panic(err)
	}
	return c
}

func TestGetOrInsertSite(t *testing.T) {
	c := testClient()
	s1, err := c.getOrInsertSite("foo")
	if err != nil {
		t.Fatal(err)
	}
	if s1.ID == 0 {
		t.Error("Expected ID to be non-zero")
	}
	if want := "foo"; s1.Name != want {
		t.Errorf("Expected Name=%q, got Name=%q", want, s1.Name)
	}
	s2, err := c.getOrInsertSite("foo")
	if s2.ID != s1.ID {
		t.Errorf("s1.ID(%d) != s2.ID(%d)", s1.ID, s2.ID)
	}
}

func TestGetSites(t *testing.T) {
	c := testClient()
	if _, err := c.getOrInsertSite("foo"); err != nil {
		t.Fatal(err)
	}
	sites, err := c.GetSites()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(sites); got != 1 {
		t.Errorf("Expected %d sites, got %d", 1, got)
	}
	if got := sites[0].Name; got != "foo" {
		t.Errorf("Expected Name=%q, got Name=%q", "foo", got)
	}
}

func TestInsertAndDeleteSites(t *testing.T) {
	c := testClient()
	var tests = []struct {
		dirCount  int
		siteCount int
		ftsCount  int
		delete    bool
	}{
		{1, 1, 1, false},
		{0, 0, 0, true},
	}
	for _, tt := range tests {
		if tt.delete {
			if err := c.DeleteSite("foo"); err != nil {
				t.Fatal(err)
			}
		} else {
			dirs := []Dir{{Path: "dir1"}}
			if err := c.Insert("foo", dirs); err != nil {
				t.Fatal(err)
			}
		}
		var count int
		if err := c.db.Get(&count, "SELECT COUNT(*) FROM site WHERE name = $1", "foo"); err != nil {
			t.Fatal(err)
		}
		if count != tt.siteCount {
			t.Errorf("Expected site row count %d, got %d", tt.siteCount, count)
		}
		if err := c.db.Get(&count, "SELECT COUNT(*) FROM dir WHERE path = $1", "dir1"); err != nil {
			t.Fatal(err)
		}
		if count != tt.dirCount {
			t.Errorf("Expected dir row count %d, got %d", tt.dirCount, count)
		}
		if err := c.db.Get(&count, "SELECT COUNT(*) FROM dir_fts WHERE path MATCH $1", "dir1"); err != nil {
			t.Fatal(err)
		}
		if count != tt.ftsCount {
			t.Errorf("Expected dir_fts row count %d, got %d", tt.ftsCount, count)
		}
	}
}

func TestSelectDirsQuery(t *testing.T) {
	var tests = []struct {
		keywords string
		site     string
		limit    int
		query    string
		args     []interface{}
	}{
		{"foo", "", 0, `SELECT site.name AS site, dir_fts.path, dir.name, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 ORDER BY site.name ASC, dir.modified DESC`, []interface{}{"foo"}},
		{"foo", "bar", 0, `SELECT site.name AS site, dir_fts.path, dir.name, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 AND site.name = $2 ORDER BY site.name ASC, dir.modified DESC`, []interface{}{"foo", "bar"}},
		{"foo", "", 10, `SELECT site.name AS site, dir_fts.path, dir.name, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 ORDER BY site.name ASC, dir.modified DESC LIMIT 10`, []interface{}{"foo"}},
	}
	for _, tt := range tests {
		query, args := selectDirsQuery(tt.keywords, tt.site, tt.limit)
		if query != tt.query || !reflect.DeepEqual(args, tt.args) {
			t.Errorf("selectDirsQuery(%q, %q, %d) => (%q, %q), want (%q, %q)", tt.keywords, tt.site, tt.limit, query, args, tt.query, tt.args)
		}
	}
}

func TestSelectDirs(t *testing.T) {
	c := testClient()
	if err := c.Insert("site1", []Dir{{Path: "/dir/foo"}, {Path: "/dir/bar"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.Insert("site2", []Dir{{Path: "/dir/foo"}, {Path: "/dir/bar"}}); err != nil {
		t.Fatal(err)
	}
	var tests = []struct {
		keywords string
		site     string
		limit    int
		out      int
	}{
		{"foo", "", 0, 2},
		{"foo", "site2", 0, 1},
		{"foo", "", 1, 1},
	}
	for _, tt := range tests {
		dirs, err := c.SelectDirs(tt.keywords, tt.site, tt.limit)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(dirs); got != tt.out {
			t.Errorf("Expected %d row(s), got %d", tt.out, got)
		}
	}
}
