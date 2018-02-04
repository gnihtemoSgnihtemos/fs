package ftp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var fieldSplitter = regexp.MustCompile(`\s+`)

type File struct {
	Path       string
	Name       string
	User       string
	Group      string
	NumEntries int
	Size       int
	Modified   time.Time
	Mode       os.FileMode
}

func parseMode(s string) (os.FileMode, error) {
	if len(s) != 10 {
		return os.FileMode(0), fmt.Errorf("invalid mode: %q", s)
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

func parseTime(now time.Time, yearOrTime, month string, day int) (time.Time, error) {
	parsedMonth, err := time.Parse("Jan", month)
	if err != nil {
		return time.Time{}, err
	}
	year := now.Year()
	hour := 0
	min := 0
	// Parse /bin/ls time format: https://cr.yp.to/ftp/list/binls.html
	// If time contains hours and minutes, the time is assumbed to be within the last 6 months
	if strings.Contains(yearOrTime, ":") {
		parts := strings.Split(yearOrTime, ":")
		if len(parts) != 2 {
			return time.Time{}, fmt.Errorf("invalid hours and minutes: %q", yearOrTime)
		}
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, err
		}
		hour = h
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, err
		}
		min = m
		// Assume time is in previous year if the month has not passed yet
		if parsedMonth.Month() > now.Month() {
			year = now.AddDate(-1, 0, 0).Year()
		}
	} else {
		y, err := strconv.Atoi(yearOrTime)
		if err != nil {
			return time.Time{}, err
		}
		year = y
	}
	return time.Date(year, parsedMonth.Month(), day, hour, min, 0, 0, time.UTC), nil
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
	parts := fieldSplitter.Split(s, 9)
	if len(parts) != 9 {
		return File{}, fmt.Errorf("invalid file format: %q", s)
	}
	fileMode, err := parseMode(parts[0])
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
	modified, err := parseTime(time.Now(), parts[7], parts[5], day)
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

func ParseFiles(path string, r io.Reader) ([]File, error) {
	var files []File
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "213") {
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
		f.Path = filepath.Join(path, f.Name)
		files = append(files, f)
	}
	return files, scanner.Err()
}

func (f *File) IsCurrentOrParent() bool {
	return f.Name == "." || f.Name == ".."
}

func (f *File) IsSymlink() bool {
	return f.Mode&os.ModeSymlink != 0
}
