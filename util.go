package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"time"
)

// isEmail returns whether the given string is an email (by using a lenient regex)
func isEmail(s string) bool {
	return matches("^.+@.+\\..+$", s)
}

func isSlackUser(s string) bool {
	return matches("^U.{8}$", s)
}

func matches(prog, input string) bool {
	re, err := regexp.Compile(prog)
	if err != nil {
		log.Printf("Error while compiling regex: %s", err.Error())
		return false
	}
	return re.Match([]byte(input))
}

func stringDefault(s, fallback string) string {
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
