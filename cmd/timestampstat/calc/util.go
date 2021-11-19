package calc

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func loadTimestamps(filePath string) (timestamps map[string]*time.Time, err error) {
	fileDesc, e := os.Open(filePath)
	if e != nil {
		err = errors.Wrapf(err, "cannot open file '%v'", filePath)
		return
	}
	defer fileDesc.Close()

	timestamps = make(map[string]*time.Time)
	scanner := bufio.NewScanner(fileDesc)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "~")
		timestamp, e := time.Parse(time.RFC3339, parts[1])
		if e != nil {
			err = errors.Wrapf(e, "invalid timestamp string '%v'", line)
			return
		}

		timestamps[parts[0]] = &timestamp
	}

	return
}

func getMin(timestamps map[string]*time.Time) *time.Time {
	if len(timestamps) == 0 {
		return nil
	}

	var min *time.Time
	for _, t := range timestamps {
		if min == nil {
			min = t
			continue
		}

		if t.UnixNano() < min.UnixNano() {
			min = t
		}
	}

	return min
}

func getMax(timestamps map[string]*time.Time) *time.Time {
	if len(timestamps) == 0 {
		return nil
	}

	var max *time.Time
	for _, t := range timestamps {
		if max == nil {
			max = t
			continue
		}

		if t.UnixNano() > max.UnixNano() {
			max = t
		}
	}

	return max
}
