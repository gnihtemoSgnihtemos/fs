package crawler

import (
	"os"
	"reflect"
	"testing"

	"github.com/martinp/ftpsc/database"
	"github.com/martinp/ftpsc/ftp"
)

type testLister struct {
}

func (l *testLister) List(path string) ([]ftp.File, error) {
	if path == "/foo/bar/baz" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "baz-regular"},
			{Name: "baz-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/bar/bax" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "bax-regular"},
			{Name: "bax-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/baz/def" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "def-regular"},
			{Name: "def-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/baz/xyz" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "xyz-regular"},
			{Name: "xyz-dir", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/bar" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "baz", Mode: os.ModeDir},
			{Name: "bax", Mode: os.ModeDir},
		}, nil
	}
	if path == "/foo/baz" {
		return []ftp.File{
			{Name: ".", Mode: os.ModeDir},
			{Name: "..", Mode: os.ModeDir},
			{Name: "def", Mode: os.ModeDir},
			{Name: "xyz", Mode: os.ModeDir},
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
	return []ftp.File{
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "foo", Mode: os.ModeDir},
	}, nil
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
		{Name: "baz", Mode: os.ModeDir},
		{Name: "bax", Mode: os.ModeDir},
		{Name: ".", Mode: os.ModeDir},
		{Name: "..", Mode: os.ModeDir},
		{Name: "def", Mode: os.ModeDir},
		{Name: "xyz", Mode: os.ModeDir},
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
		{Path: "/foo/bar", Name: "bar"},
		{Path: "/foo/bar/baz", Name: "baz"},
	}
	want := database.Dir{Path: "/foo/bar/baz", Name: "baz"}
	got := filterFiles(files)
	if len(got) == 0 {
		t.Fatal("expected non-zero length")
	}
	if got[0].Name != want.Name {
		t.Errorf("got Name=%q, want Name=%q", got[0].Name, want.Name)
	}
	if got[0].Path != want.Path {
		t.Errorf("got Path=%q, want Path=%q", got[0].Path, want.Path)
	}
}
