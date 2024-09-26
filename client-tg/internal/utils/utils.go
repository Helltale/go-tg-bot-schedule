package utils

import (
	"fmt"
	"time"
)

func FormatWorkTime(start, end string) (string, error) {

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return "", err
	}
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%02d:%02d - %02d:%02d", startTime.Hour(), startTime.Minute(), endTime.Hour(), endTime.Minute()), nil
}
