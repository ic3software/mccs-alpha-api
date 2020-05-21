package util

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jinzhu/now"
)

// ParseTime parses string into time.
func ParseTime(s string) time.Time {
	if s == "" || s == "1-01-01 00:00:00 UTC" {
		return time.Time{}
	}

	parseUnixTime, err := parseAsUnixTime(s)
	if err == nil {
		return parseUnixTime
	}

	now.TimeFormats = append(now.TimeFormats,
		"2 January 2006",
		"2 January 2006 3:04 PM",
		"2006-01-02 03:04:05 MST",
	)

	t, err := now.ParseInLocation(time.UTC, s)
	if err != nil {
		log.Printf("[ERROR] ParseTime failed: %+v", err)
		return time.Time{}
	}
	return t
}

func parseAsUnixTime(input string) (time.Time, error) {
	i, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return time.Time{}, errors.New("error while pasing input as the unit time")
	}
	return time.Unix(i, 0), nil
}

// FormatTime formats time in UK format.
func FormatTime(t time.Time) string {
	tt := t.UTC()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d UTC",
		tt.Year(), tt.Month(), tt.Day(),
		tt.Hour(), tt.Minute(), tt.Second())
}
