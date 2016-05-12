package ftp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type File struct {
	Name       string
	User       string
	Group      string
	NumEntries int
	Size       int
	Modified   time.Time
	Mode       os.FileMode
}

func ParseFileMode(s string) (os.FileMode, error) {
	if len(s) != 10 {
		return os.FileMode(0), fmt.Errorf("length must be 10")
	}
	var mode os.FileMode
	for i, c := range s {
		switch c {
		case 'd':
			mode += os.ModeDir
		case 'l':
			mode += os.ModeSymlink
		case 'r', 'w', 'x':
			mode += 1 << uint(9-i)
		}
	}
	return mode, nil
}

func ParseTime(yearOrTime, month string, day int) (time.Time, error) {
	m, err := time.Parse("Jan", month)
	if err != nil {
		return time.Time{}, err
	}
	year := time.Now().Year()
	hour := 0
	min := 0
	if strings.Contains(yearOrTime, ":") {
		parts := strings.Split(yearOrTime, ":")
		if len(parts) != 2 {
			return time.Time{}, fmt.Errorf("could not parse hours and minutes from %s", yearOrTime)
		}
		_hour, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, err
		}
		hour = _hour
		_min, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, err
		}
		min = _min
	} else {
		_year, err := strconv.Atoi(yearOrTime)
		if err != nil {
			return time.Time{}, err
		}
		year = _year
	}
	return time.Date(year, m.Month(), day, hour, min, 0, 0, time.UTC), nil
}

func ParseFile(s string) (File, error) {
	// 0 = filemode
	// 1 = num of entries in directory
	// 2 = user
	// 3 = group
	// 4 = size
	// 5 = month
	// 6 = day
	// 7 = time or year
	// 8 = filename
	parts := regexp.MustCompile("\\s+").Split(s, 9)
	if len(parts) != 9 {
		return File{}, fmt.Errorf("failed to parse file: %s", s)
	}
	fileMode, err := ParseFileMode(parts[0])
	if err != nil {
		return File{}, err
	}
	numEntries, err := strconv.Atoi(parts[1])
	if err != nil {
		return File{}, err
	}
	user := parts[2]
	group := parts[3]
	size, err := strconv.Atoi(parts[4])
	if err != nil {
		return File{}, err
	}
	day, err := strconv.Atoi(parts[6])
	if err != nil {
		return File{}, err
	}
	modified, err := ParseTime(parts[7], parts[5], day)
	if err != nil {
		return File{}, err
	}
	return File{
		Name:       parts[8],
		User:       user,
		Group:      group,
		Size:       size,
		NumEntries: numEntries,
		Modified:   modified,
		Mode:       fileMode,
	}, nil
}

func ParseFiles(r io.Reader) ([]File, error) {
	files := []File{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) < 9 {
			continue
		}
		f, err := ParseFile(line)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func (f *File) IsSymlink() bool {
	return f.Mode&os.ModeSymlink != 0
}

func (f *File) Age(since time.Time) time.Duration {
	return since.Round(time.Second).Sub(f.Modified)
}
