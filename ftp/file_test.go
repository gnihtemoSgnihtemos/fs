package ftp

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseFileMode(t *testing.T) {
	var tests = []struct {
		in string
	}{
		{"drwxrwxrwx"},
		{"lrwxrwxrwx"},
		{"-rwxrwxrwx"},
		{"-rwxr--r--"},
		{"-r-xr-xr-x"},
	}
	for _, tt := range tests {
		out, err := parseFileMode(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if s := strings.ToLower(out.String()); s != tt.in {
			t.Errorf("ParseFileMode(%q) => %q, want %q", tt.in, s, tt.in)
		}
	}
}

func TestParseTime(t *testing.T) {
	var tests = []struct {
		day        int
		month      string
		yearOrTime string
		out        time.Time
	}{
		{15, "Jan", "2014", time.Date(2014, 1, 15, 0, 0, 0, 0, time.UTC)},
		{7, "Oct", "23:14", time.Date(time.Now().Year(), 10, 7, 23, 14, 0, 0, time.UTC)},
		{21, "Jul", "05:32", time.Date(time.Now().Year(), 7, 21, 5, 32, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		rv, err := parseTime(tt.yearOrTime, tt.month, tt.day)
		if err != nil {
			t.Fatal(err)
		}
		if !rv.Equal(tt.out) {
			t.Errorf("ParseTime(%d, %q, %q) => %s, want %s", tt.day, tt.month, tt.yearOrTime, rv, tt.out)
		}
	}
}

func TestParseFile(t *testing.T) {
	var tests = []struct {
		in  string
		out File
	}{
		{"drwxrwxrwx   3 foo   bar       4096 Jul 25   2014 dir with spaces",
			File{
				Name: "dir with spaces",
				User: "foo", Group: "bar",
				NumEntries: 3,
				Size:       4096,
				Modified:   time.Date(2014, 7, 25, 0, 0, 0, 0, time.UTC),
				Mode:       os.FileMode(os.ModeDir + 0777),
			},
		},
		{"drwxrwxrwx   3 bax   baz     131072 Jan 19  23:14 dir",
			File{
				Name: "dir",
				User: "bax", Group: "baz",
				NumEntries: 3,
				Size:       131072,
				Modified:   time.Date(time.Now().Year(), 1, 19, 23, 14, 0, 0, time.UTC),
				Mode:       os.FileMode(os.ModeDir + 0777),
			},
		},
	}
	for _, tt := range tests {
		rv, err := ParseFile(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rv, tt.out) {
			t.Errorf("ParseFile(%q) => %+v, want %+v", tt.in, rv, tt.out)
		}
	}
}

func TestParseFiles(t *testing.T) {
	message := `213- status of -al .:
total 1488
drwxr-xr-x  19 foo   bar       4096 Dec 23 13:00 .
drwxr-xr-x  16 foo   bar       4096 May  2  2015 ..
drwxrwxrwx   3 foo   bar       4096 Jul  3  2014 dir1
drwxrwxrwx  19 foo   bar       4096 Apr 10 08:41 dir2
drwxrwxrwx  72 foo   bar      94208 May 15 01:03 dir3
213 End of Status`
	files, err := ParseFiles("/", strings.NewReader(message))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(files); got != 5 {
		t.Errorf("got %d, want %d", got, 5)
	}
}

func TestParseFilesSkipsLinesWithReturnCodes(t *testing.T) {
	message := "213- status of -al /a b c d e f:"
	files, err := ParseFiles("/", strings.NewReader(message))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(files); got != 0 {
		t.Errorf("got %d, want %d", got, 0)
	}
}

func TestIsCurrentOrParent(t *testing.T) {
	var tests = []struct {
		in  string
		out bool
	}{
		{".", true},
		{"..", true},
		{"foo", false},
	}
	for _, tt := range tests {
		got := (&File{Name: tt.in}).IsCurrentOrParent()
		if got != tt.out {
			t.Errorf("File{Name: %q}.IsCurrentOrParent() => %t, want %t", tt.in, got, tt.out)
		}
	}
}

func TestIsSymlink(t *testing.T) {
	var tests = []struct {
		in  File
		out bool
	}{
		{File{Mode: os.ModeSymlink}, true},
		{File{Mode: os.ModeDir}, false},
		{File{}, false},
	}
	for i, tt := range tests {
		got := tt.in.IsSymlink()
		if got != tt.out {
			t.Errorf("File[%d].IsSymlink() => %t, want %t", i, got, tt.out)
		}
	}
}
