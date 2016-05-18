package cmd

import (
	"testing"

	"github.com/martinp/fs/crawler"
	"github.com/martinp/fs/database"
)

func TestDifference(t *testing.T) {
	sites := []database.Site{
		{Name: "foo"},
		{Name: "bar"},
		{Name: "baz"},
	}
	configSites := []crawler.Site{{Name: "foo"}}
	diff := difference(sites, configSites)
	if got := len(diff); got != 2 {
		t.Fatalf("Expected 2 sites, got %d", got)
	}
	if want := "bar"; diff[0].Name != want {
		t.Errorf("Site with Name=%q", want)
	}
	if want := "baz"; diff[1].Name != want {
		t.Errorf("Site with Name=%q", want)
	}
}
