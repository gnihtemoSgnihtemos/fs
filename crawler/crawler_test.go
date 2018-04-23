package crawler

import (
	"fmt"
	"os"
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
	case "/dir2/Dir2-1":
		return []ftp.File{
			{Name: "file2-1-1"}, // Regular, but should not decide max depth
		}, nil
	case "/dir2/_dir2-2":
		return []ftp.File{
			{Name: "dir2-2-1", Mode: os.ModeDir},
		}, nil
	case "/dir2":
		return []ftp.File{
			{Name: "Dir2-1", Mode: os.ModeDir}, // Names starting with '_' should sort before uppercase chars
			{Name: "_dir2-2", Mode: os.ModeDir},
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
		{Name: "Dir2-1", Mode: os.ModeDir},
		{Name: "dir2-2-1", Mode: os.ModeDir},
	}
	got, err := walkShallow(&fakeLister{}, "/", -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != len(got) {
		t.Fatal("Expected equal length")
	}
	for i := range want {
		if want[i].Name != got[i].Name {
			t.Errorf("want Name=%s, got Name=%s", want[i].Name, got[i].Name)
		}
		if want[i].Mode != got[i].Mode {
			t.Errorf("want Mode=%d, got Mode=%d", want[i].Mode, got[i].Mode)
		}
	}
}

func TestMakeDirs(t *testing.T) {
	files := []ftp.File{
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
