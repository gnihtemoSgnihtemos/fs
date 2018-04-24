package crawler

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/mpolden/fs/database"
	"github.com/mpolden/fs/ftp"
)

type fakeLister struct{}

func (l *fakeLister) filterFiles(files []ftp.File) []ftp.File {
	return filterFiles(files, []string{"_foo", "_bar"}, true)
}

func (l *fakeLister) list(path string) ([]ftp.File, error) {
	switch path {
	case "/dir2/_dir2-2/dir2-2-1":
		return []ftp.File{
			{Name: "file2-2-1"}, // Regular, decides depth
		}, nil
	case "/dir2/ADir2-3":
		return []ftp.File{
			{Name: "file2-1-1"}, // Regular
		}, nil
	case "/dir2/Dir2-1":
		return []ftp.File{
			{Name: "file2-1-1"}, // Regular
		}, nil
	case "/dir2/_dir2-2":
		return []ftp.File{
			{Name: "dir2-2-1", Mode: os.ModeDir},
		}, nil
	case "/dir2":
		return []ftp.File{
			{Name: "Dir2-1", Mode: os.ModeDir}, // Names starting with '_' should sort before uppercase chars
			{Name: "_dir2-2", Mode: os.ModeDir},
			{Name: "ADir2-3", Mode: os.ModeDir},
		}, nil
	case "/dir1/dir1-1/dir1-1-1":
		return []ftp.File{
			{Name: "file1-1-1-1"}, // Regular
			{Name: "dir1-1-1-1", Mode: os.ModeDir},
		}, nil
	case "/dir1/dir1-2/dir1-2-1":
		return []ftp.File{
			{Name: "file1-2-1-1"},
			{Name: "dir1-2-1-1", Mode: os.ModeDir},
		}, nil
	case "/dir1/dir1-2":
		return []ftp.File{
			{Name: "dir1-2-1", Mode: os.ModeDir},
			{Name: "dir1-2-2", Mode: os.ModeDir},
			// Ignored:
			{Name: "_foo", Mode: os.ModeDir},
			{Name: "_bar", Mode: os.ModeDir},
			{Name: "symlink1-2-1", Mode: os.ModeSymlink},
		}, nil
	case "/dir1/dir1-1":
		return []ftp.File{
			{Name: "dir1-1-1", Mode: os.ModeDir},
			{Name: "dir1-1-2", Mode: os.ModeDir},
		}, nil
	case "/dir1":
		return []ftp.File{
			{Name: "dir1-1", Mode: os.ModeDir},
			{Name: "dir1-2", Mode: os.ModeDir},
		}, nil
	case "/":
		return []ftp.File{
			{Name: "dir1", Mode: os.ModeDir},
			{Name: "dir2", Mode: os.ModeDir},
		}, nil
	}
	return nil, fmt.Errorf("unknown path: %s", path)
}

func TestWalk(t *testing.T) {
	want := []ftp.File{
		{Name: "dir1", Mode: os.ModeDir},
		{Name: "dir2", Mode: os.ModeDir},
		{Name: "dir1-1", Mode: os.ModeDir},
		{Name: "dir1-2", Mode: os.ModeDir},
		{Name: "dir1-1-1", Mode: os.ModeDir},
		{Name: "dir1-1-2", Mode: os.ModeDir},
		{Name: "dir1-2-1", Mode: os.ModeDir},
		{Name: "dir1-2-2", Mode: os.ModeDir},
		{Name: "_dir2-2", Mode: os.ModeDir},
		{Name: "ADir2-3", Mode: os.ModeDir},
		{Name: "Dir2-1", Mode: os.ModeDir},
		{Name: "dir2-2-1", Mode: os.ModeDir},
	}
	got, err := walk(&fakeLister{}, "/", -1)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		other := ftp.File{}
		if i < len(got) {
			other = got[i]
		}
		if want[i].Name != other.Name {
			t.Errorf("want Name=%s, got Name=%s", want[i].Name, other.Name)
		}
		if want[i].Mode != other.Mode {
			t.Errorf("want Mode=%d, got Mode=%d", want[i].Mode, other.Mode)
		}
	}
}

func TestSortFiles(t *testing.T) {
	got := []ftp.File{
		{Name: "_C"},
		{Name: "_b"},
		{Name: "_A"},
		{Name: "_B"},
		{Name: "B"},
		{Name: "A"},
		{Name: "_c"},
		{Name: "C"},
		{Name: "_a"},
		{Name: ""},
	}
	sortFiles(got)
	want := []ftp.File{
		{Name: "_A"},
		{Name: "_B"},
		{Name: "_C"},
		{Name: "_a"},
		{Name: "_b"},
		{Name: "_c"},
		{Name: ""},
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %+v, got %+v", want, got)
	}
}

func TestToDirs(t *testing.T) {
	files := []ftp.File{{Path: "/foo", Name: "foo"}}
	want := database.Dir{Path: "/foo"}
	got := toDirs(files)
	if len(got) == 0 {
		t.Fatal("expected non-zero length")
	}
	if got[0].Path != want.Path {
		t.Errorf("want Path=%s, got Path=%s", want.Path, got[0].Path)
	}
}
