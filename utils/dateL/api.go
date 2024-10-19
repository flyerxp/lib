package dateL

import "time"

func SplitDateRange(startDate time.Time, endDate time.Time, step int) [][]time.Time {
	var result [][]time.Time
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		nextDate := currentDate.AddDate(0, 0, step)
		if nextDate.After(endDate) {
			result = append(result, []time.Time{currentDate, endDate})
			break
		} else {
			result = append(result, []time.Time{currentDate, nextDate})
			currentDate = nextDate
		}
	}
	return result
}
