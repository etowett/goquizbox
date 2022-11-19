package util

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	timeLayout            = "2006/01/02 15:04:05"
	timeandtimezonelayout = "2006/01/02 15:04:05-07:00"
)

func ParseTime(
	timeString string,
) (time.Time, error) {

	parsedTime, err := time.ParseInLocation(timeLayout, timeString, time.UTC)
	if err != nil {

		parsedTime, err = time.ParseInLocation(timeandtimezonelayout, timeString, time.UTC)
		if err != nil {
			return parsedTime, fmt.Errorf("could not parse time in location: %v", err)
		}
	}

	return parsedTime, nil
}

func FormatTime(
	timeToFormat time.Time,
) string {

	formattedTime := timeToFormat.In(time.UTC).Format(timeLayout)

	return formattedTime
}

func BackoffRetryDelay(retry int) time.Duration {

	duration := backOff(retry)

	rand.Seed(time.Now().UnixNano())
	durationWithJitter := rand.Intn(duration) + duration

	return time.Duration(durationWithJitter) * time.Second
}

func backOff(retry int) int {

	increments := []int{5, 7, 11, 13, 17, 19, 23, 29, 31, 37}

	increment := retry % 10

	if increment > 9 {
		increment = 0
	}

	factor := increments[increment]

	return factor * 60
}

func UnixMillis(t time.Time) int64 {
	return t.UnixNano() / 1e6
}
