package main

import (
	"fmt"
	"math"
	"regexp"
	"time"
)

func isEmail(s string) bool {
	return regexp.MustCompile(`^.+@.+\..+$`).MatchString(s)
}

func isSlackUser(s string) bool {
	return regexp.MustCompile(`^U.{8}$`).MatchString(s)
}

func stringOrDefault(s, fallback string) string {
	if s == "" {
		return fallback
	} else {
		return s
	}
}

// getTimestamp returns timestamps in the format the slack API uses
func getTimestamp() string {
	nanos := time.Now().UnixNano()
	ts := fmt.Sprintf("%.6f", float64(nanos)/math.Pow(10, 9))
	return ts
}
