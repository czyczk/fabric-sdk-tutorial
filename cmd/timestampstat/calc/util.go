package calc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func loadStartEndTimestamps(startLogFilePath, endLogFilePath string) (startTimestamps map[string]*time.Time, endTimestamps map[string]*time.Time, err error) {
	startLogFd, e := os.Open(startLogFilePath)
	if e != nil {
		err = errors.Wrapf(err, "cannot open file '%v'", startLogFilePath)
		return
	}
	defer startLogFd.Close()

	endLogFd, e := os.Open(endLogFilePath)
	if e != nil {
		err = errors.Wrapf(err, "cannot open file '%v'", endLogFilePath)
		return
	}
	defer endLogFd.Close()

	startTimestamps = make(map[string]*time.Time)
	endTimestamps = make(map[string]*time.Time)

	/* Process the start log:
	 * For each "${loggerID}~${timestamp}~${isSuccess}" line in the log, store the mappings of
	 * ${loggerID} -> ${timestamp} in the map.
	 * In the start log, we assume ${isSuccess} is always `true`.
	 */
	scanner := bufio.NewScanner(startLogFd)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "~")
		if len(parts) != 3 {
			err = errors.Wrapf(e, "cannot parse line '%v'", line)
			return
		}

		timestamp, e := time.Parse(time.RFC3339, parts[1])
		if e != nil {
			err = errors.Wrapf(e, "invalid timestamp string in line '%v'", line)
			return
		}

		startTimestamps[parts[0]] = &timestamp
	}

	/* Process the end log:
	 * For each "${loggerID}~${timestamp}~${isSuccess}" line in the log,
	 *   1. Store the mappings of ${loggerID} -> ${timestamp} in the map if ${isSuccess} is `true`.
	 *   2. Delete the entry from the start map if ${isSuccess} is `false`.
	 */
	scanner = bufio.NewScanner(endLogFd)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "~")
		if len(parts) != 3 {
			err = errors.Wrapf(e, "cannot parse line '%v'", line)
			return
		}

		timestamp, e := time.Parse(time.RFC3339, parts[1])
		if e != nil {
			err = errors.Wrapf(e, "invalid timestamp in line '%v'", line)
			return
		}

		isSuccess, e := strconv.ParseBool(parts[2])
		if e != nil {
			err = errors.Wrapf(e, "invalid bool value in line '%v'", line)
			return
		}

		if isSuccess {
			endTimestamps[parts[0]] = &timestamp
		} else {
			log.Infof("Logger ID '%v' failed on an action which will be ignored.", parts[0])
			delete(startTimestamps, parts[0])
		}
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
