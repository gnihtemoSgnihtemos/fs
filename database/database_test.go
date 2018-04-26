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

func TestSelectSites(t *testing.T) {
	c := testClient()
	if err := c.Insert("foo", nil); err != nil {
		t.Fatal(err)
	}
	sites, err := c.SelectSites()
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

func TestInsert(t *testing.T) {
	c := testClient()
	var tests = []struct {
		site string
		dirs []Dir
	}{
		{"foo", []Dir{{Path: "dir1"}}},
		{"foo", []Dir{{Path: "dir2"}, {Path: "dir3"}}}, // Overwrites previous insert
		{"bar", nil},
	}
	for _, tt := range tests {
		if err := c.Insert(tt.site, tt.dirs); err != nil {
			t.Fatal(err)
		}

		var site Site
		if err := c.db.Get(&site, "SELECT * FROM site WHERE name = $1", tt.site); err != nil {
			t.Fatal(err)
		}
		if site.Name != tt.site {
			t.Fatalf("want Name=%s, got %s", tt.site, site.Name)
		}

		var dirs []Dir
		if err := c.db.Select(&dirs, "SELECT path, modified FROM dir WHERE site_id = $1", site.ID); err != nil {
			t.Fatal(err)
		}
		if len(dirs) != len(tt.dirs) {
			t.Fatalf("want %d dirs, got %d", len(tt.dirs), len(dirs))
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
		{"foo", "", 0, `SELECT site.name AS site, dir_fts.path, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 ORDER BY site.name ASC, dir.modified DESC`, []interface{}{"foo"}},
		{"foo", "bar", 0, `SELECT site.name AS site, dir_fts.path, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 AND site.name = $2 ORDER BY site.name ASC, dir.modified DESC`, []interface{}{"foo", "bar"}},
		{"foo", "", 10, `SELECT site.name AS site, dir_fts.path, dir.modified FROM dir_fts
INNER JOIN dir ON dir_fts.id = dir.id
INNER JOIN site ON dir_fts.site_id = site.id
WHERE dir_fts.path MATCH $1 ORDER BY site.name ASC, dir.modified DESC LIMIT 10`, []interface{}{"foo"}},
	}
	for _, tt := range tests {
		query, args := selectDirsQuery(tt.keywords, tt.site, "site.name ASC, dir.modified DESC", tt.limit)
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
		dirs, err := c.SelectDirs(tt.keywords, tt.site, "", tt.limit)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(dirs); got != tt.out {
			t.Errorf("Expected %d row(s), got %d", tt.out, got)
		}
	}
}

func TestOrderByClause(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err bool
	}{
		{"foo:bar", "", true},
		{"", "", false},
		{"foo", "foo", false},
		{"foo:desc", "foo DESC", false},
	}
	for _, tt := range tests {
		got, err := OrderByClause(tt.in)
		if !tt.err && err != nil {
			t.Fatal(err)
		}
		if got != tt.out {
			t.Errorf("orderByClause(%q) => %q, want %q", tt.in, got, tt.out)
		}
	}
}

func TestOrderByClauses(t *testing.T) {
	in := []string{"foo", "bar:desc"}
	got, err := OrderByClauses(in)
	if err != nil {
		t.Fatal(err)
	}
	want := "foo, bar DESC"
	if got != want {
		t.Errorf("orderByClauses(%q) => %q, want %q", in, got, want)
	}
}
