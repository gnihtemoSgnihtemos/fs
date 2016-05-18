package database

import "testing"

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
	if err := c.insertSite("foo"); err != nil {
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
		delete    bool
	}{
		{1, 1, false},
		{0, 0, true},
	}
	for _, tt := range tests {
		if tt.delete {
			if err := c.DeleteSite("foo"); err != nil {
				t.Fatal(err)
			}
		} else {
			dirs := []Dir{{Name: "dir1"}}
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
		if err := c.db.Get(&count, "SELECT COUNT(*) FROM dir WHERE name = $1", "dir1"); err != nil {
			t.Fatal(err)
		}
		if count != tt.dirCount {
			t.Errorf("Expected dir row count %d, got %d", tt.dirCount, count)
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
	dirs, err := c.SelectDirs("foo")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(dirs); got != 2 {
		t.Errorf("Expected 2 rows, got %d", got)
	}
	dirs, err = c.SelectDirsSite("site2", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(dirs); got != 1 {
		t.Errorf("Expected 1 rows, got %d", got)
	}
}
