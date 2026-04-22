package wela

import (
	"fmt"
	"strings"
	"time"
)

func ParseISODateBKK(s string) (time.Time, error) {
	return parseLayout(s, LayoutISODate, bangkokLoc)
}

func ParseISODateTimeBKK(s string) (time.Time, error) {
	return parseLayout(s, LayoutISODateTime, bangkokLoc)
}

func ParseISODateUTC(s string) (time.Time, error) {
	return parseLayout(s, LayoutISODate, time.UTC)
}

func ParseISODateTimeUTC(s string) (time.Time, error) {
	return parseLayout(s, LayoutISODateTime, time.UTC)
}

func ParseRFC3339Any(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %q with layout %q: %w", s, time.RFC3339, err)
	}
	return t, nil
}

func ParseRFC3339BKK(s string) (time.Time, error) {
	t, err := ParseRFC3339Any(s)
	if err != nil || t.IsZero() {
		return t, err
	}
	return t.In(bangkokLoc), nil
}

func ParseThaiDateBKK(s string) (time.Time, error) {
	s = normalizeSpaces(s)
	if s == "" {
		return time.Time{}, nil
	}

	day, month, year, err := splitThaiDate(s, "ParseThaiDateBKK")
	if err != nil {
		return time.Time{}, err
	}

	t := time.Date(YearBEtoCE(year), time.Month(month), day, 0, 0, 0, 0, bangkokLoc)
	if t.Day() != day || t.Month() != time.Month(month) {
		return time.Time{}, fmt.Errorf("ParseThaiDateBKK: invalid calendar date %q", s)
	}

	return t, nil
}

func ParseThaiDateTimeBKK(s string) (time.Time, error) {
	s = normalizeSpaces(s)
	if s == "" {
		return time.Time{}, nil
	}

	parts := strings.SplitN(s, " ", 2)
	day, month, year, err := splitThaiDate(parts[0], "ParseThaiDateTimeBKK")
	if err != nil {
		return time.Time{}, err
	}

	hour, min, sec := 0, 0, 0
	if len(parts) == 2 {
		hour, min, sec, err = parseTimePart(parts[1], "ParseThaiDateTimeBKK")
		if err != nil {
			return time.Time{}, err
		}
	}

	t := time.Date(YearBEtoCE(year), time.Month(month), day, hour, min, sec, 0, bangkokLoc)
	if t.Day() != day || t.Month() != time.Month(month) {
		return time.Time{}, fmt.Errorf("ParseThaiDateTimeBKK: invalid calendar date %q", parts[0])
	}

	return t, nil
}

func ParseDMYBKK(s string) (time.Time, error) {
	return parseLayout(s, LayoutDMY, bangkokLoc)
}

// ParseDateBKKAuto รองรับหลายรูปแบบยอดนิยม:
// - 2006-01-02
// - 2006-01-02 15:04:05
// - 02/01/2006
// - 02/01/2568
// - 02/01/2568 15:04
// - 02/01/2568 15:04:05
// - RFC3339
func ParseDateBKKAuto(s string) (time.Time, error) {
	s = normalizeSpaces(s)
	if s == "" {
		return time.Time{}, nil
	}

	parsers := []func(string) (time.Time, error){
		ParseISODateTimeBKK,
		ParseISODateBKK,
		ParseThaiDateTimeBKK,
		ParseThaiDateBKK,
		ParseDMYBKK,
		ParseRFC3339BKK,
	}

	for _, fn := range parsers {
		t, err := fn(s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("ParseDateBKKAuto: unsupported date format %q", s)
}
