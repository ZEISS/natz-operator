package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseExpiry parses a string into an expiry time.
// nolint:gocyclo
func ParseExpiry(s string) (int64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	if strings.Contains(s, ".") {
		return 0, fmt.Errorf("start/expiry must be an integer value: %v", s)
	}

	t, err := time.Parse("2006-01-02 15:04:05 UTC", s)
	if err == nil {
		return t.Unix(), nil
	}

	re := regexp.MustCompile(`(\d){4}-(\d){2}-(\d){2}`)
	if re.MatchString(s) {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return 0, err
		}
		return t.Unix(), nil
	}

	re = regexp.MustCompile(`(?P<count>-?\d+)(?P<qualifier>[mhdMyw])`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, err
		}
		count := int(v)
		if count == 0 {
			return 0, nil
		}
		dur := time.Duration(count)
		now := time.Now()

		switch m[2] {
		case "m":
			// nolint:durationcheck
			return now.Add(dur * time.Minute).Unix(), nil
		case "h":
			// nolint:durationcheck
			return now.Add(dur * time.Hour).Unix(), nil
		case "d":
			return now.AddDate(0, 0, count).Unix(), nil
		case "w":
			return now.AddDate(0, 0, 7*count).Unix(), nil
		case "M":
			return now.AddDate(0, count, 0).Unix(), nil
		case "y":
			return now.AddDate(count, 0, 0).Unix(), nil
		default:
			return 0, fmt.Errorf("unknown interval %q in %q", m[2], m[0])
		}
	}

	return 0, fmt.Errorf("couldn't parse expiry: %v", s)
}
