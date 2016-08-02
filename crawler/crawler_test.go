package crawler

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/martinp/fs/database"
	"github.com/martinp/fs/ftp"
)

type testLister struct {
}

func (l *testLister) FilterFiles(files []ftp.File) []ftp.File {
	return filterFiles(files, []string{"_baz", "_foo"}, true)
}

func (l *testLister) List(path string) ([]ftp.File, error) {
	if path == "/foo/bar/bar1" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "bar1-regular"},
			{Name: "bar1-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/baz/baz1" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "baz1-regular"},
			{Name: "baz1-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/bar" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "bar1", Mode: os.ModeDir},
			{Name: "bar2", Mode: os.ModeDir},
			{Name: "_foo", Mode: os.ModeDir},
			{Name: "_baz", Mode: os.ModeDir},
			{Name: "_bax", Mode: os.ModeSymlink},
		}, nil
	}
	if path == "/foo/baz" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "baz1", Mode: os.ModeDir},
			{Name: "baz2", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "bar", Mode: os.ModeDir},
			{Name: "baz", Mode: os.ModeDir},
		}, nil
	}
	if path == "/" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "foo", Mode: os.ModeDir},
		}, nil
	}
	return nil, fmt.Errorf("unknown path: %s", path)
}

func TestWalk(t *testing.T) {
	want := []ftp.File{
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "foo", Mode: os.ModeDir},
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "bar", Mode: os.ModeDir},
		{Name: "baz", Mode: os.ModeDir},
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "bar1", Mode: os.ModeDir},
		{Name: "bar2", Mode: os.ModeDir},
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "baz1", Mode: os.ModeDir},
		{Name: "baz2", Mode: os.ModeDir},
	}
	got, err := walkShallow(&testLister{}, "/", -1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkDirs(%q) => %+v, want %+v", "/", got, want)
	}
}

func TestMakeDirs(t *testing.T) {
	files := []ftp.File{
		{Path: "/foo", Name: "."},
		{Path: "/foo", Name: ".."},
		{Path: "/foo", Name: "foo"},
	}
	want := database.Dir{Path: "/foo"}
	got := makeDirs(files)
	if len(got) == 0 {
		t.Fatal("expected non-zero length")
	}
	if got[0].Path != want.Path {
		t.Errorf("got Path=%q, want Path=%q", got[0].Path, want.Path)
	}
}
