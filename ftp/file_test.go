package ftp

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return dt(year, month, day, 0, 0, 0)
}

func dt(year int, month time.Month, day, hour, minute, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

func TestParseMode(t *testing.T) {
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
		out, err := parseMode(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if s := strings.ToLower(out.String()); s != tt.in {
			t.Errorf("ParseMode(%q) => %q, want %q", tt.in, s, tt.in)
		}
	}
}

func TestParseTime(t *testing.T) {
	var tests = []struct {
		day        int
		month      string
		yearOrTime string
		out        time.Time
		now        time.Time
	}{
		{15, "Jan", "2014", date(2014, 1, 15), time.Now()},
		{7, "Oct", "23:14", dt(2016, 10, 7, 23, 14, 0), date(2016, 11, 1)},
		{21, "Jul", "05:32", dt(2016, 7, 21, 5, 32, 0), date(2016, 11, 1)},
		{10, "Dec", "09:24", dt(2017, 12, 10, 9, 24, 0), date(2018, 1, 1)},
		{10, "Jan", "09:24", dt(2018, 1, 10, 9, 24, 0), date(2018, 1, 1)},
	}
	for _, tt := range tests {
		rv, err := parseTime(tt.now, tt.yearOrTime, tt.month, tt.day)
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
				Modified:   date(2014, 7, 25),
				Mode:       os.FileMode(os.ModeDir + 0777),
			},
		},
		{"drwxrwxrwx   3 bax   baz     131072 Jan 19  23:14 dir",
			File{
				Name: "dir",
				User: "bax", Group: "baz",
				NumEntries: 3,
				Size:       131072,
				Modified:   dt(time.Now().Year(), 1, 19, 23, 14, 0),
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
	if want, got := 3, len(files); got != want {
		t.Errorf("got %d, want %d", got, want)
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
