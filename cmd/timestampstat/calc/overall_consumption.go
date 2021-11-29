package calc

import (
	"fmt"
	"time"
)

func CalcTimeConsumptions(filePathBefore, filePathAfter string) (overallConsumption time.Duration, avgConsumption time.Duration, err error) {
	// Parse timestamps from the files
	timestampsBefore, timestampsAfter, err := loadStartEndTimestamps(filePathBefore, filePathAfter)
	if err != nil {
		return
	}

	// Check if the numbers of timestamps from the "before" file and the "after" file are the same
	lenTimestampsBefore, lenTimestampsAfter := len(timestampsBefore), len(timestampsAfter)
	if lenTimestampsBefore != lenTimestampsAfter {
		err = fmt.Errorf("timestamp counts not equal, before:after=%v:%v", lenTimestampsBefore, lenTimestampsAfter)
		return
	}

	if lenTimestampsBefore == 0 {
		err = fmt.Errorf("file is empty")
		return
	}

	// Calc overall consumption
	minBefore := getMin(timestampsBefore)
	maxAfter := getMax(timestampsAfter)
	overallConsumption = maxAfter.Sub(*minBefore)

	// Calc avg consumption
	consumptionSum := time.Duration(0)
	for id, tBefore := range timestampsBefore {
		tAfter, ok := timestampsAfter[id]
		if !ok {
			err = fmt.Errorf("timestamp for worker id %v not found in \"after\" file", id)
			return
		}
		consumptionSum += tAfter.Sub(*tBefore)
	}

	avgConsumption = time.Duration(int64(consumptionSum) / int64(lenTimestampsBefore))
	return
}
