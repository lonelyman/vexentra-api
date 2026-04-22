package wela

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseLayout(s, layout string, loc *time.Location) (time.Time, error) {
	s = normalizeSpaces(s)
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.ParseInLocation(layout, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %q with layout %q: %w", s, layout, err)
	}
	return t, nil
}

func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func splitThaiDate(s, caller string) (day, month, year int, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, 0, fmt.Errorf("%s: empty string", caller)
	}

	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("%s: invalid format %q (expect DD/MM/YYYY)", caller, s)
	}

	day, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%s: invalid day %q", caller, parts[0])
	}

	month, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%s: invalid month %q", caller, parts[1])
	}

	year, err = strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%s: invalid year %q", caller, parts[2])
	}

	if month < 1 || month > 12 {
		return 0, 0, 0, fmt.Errorf("%s: month out of range %d", caller, month)
	}
	if day < 1 || day > 31 {
		return 0, 0, 0, fmt.Errorf("%s: day out of range %d", caller, day)
	}

	return day, month, year, nil
}

func parseTimePart(s, caller string) (hour, min, sec int, err error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, 0, 0, fmt.Errorf("%s: invalid time %q (expect HH:MM or HH:MM:SS)", caller, s)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, 0, fmt.Errorf("%s: invalid hour %q", caller, parts[0])
	}

	min, err = strconv.Atoi(parts[1])
	if err != nil || min < 0 || min > 59 {
		return 0, 0, 0, fmt.Errorf("%s: invalid minute %q", caller, parts[1])
	}

	if len(parts) == 3 {
		sec, err = strconv.Atoi(parts[2])
		if err != nil || sec < 0 || sec > 59 {
			return 0, 0, 0, fmt.Errorf("%s: invalid second %q", caller, parts[2])
		}
	}

	return hour, min, sec, nil
}

func formatThaiDate(t time.Time, shortMonth bool, beYear bool, withWeekday bool) string {
	if t.IsZero() {
		return ""
	}

	monthName := thaiMonthsFull[int(t.Month())]
	if shortMonth {
		monthName = thaiMonthsShort[int(t.Month())]
	}

	year := t.Year()
	if beYear {
		year = YearCEtoBE(year)
	}

	body := fmt.Sprintf("%d %s %d", t.Day(), monthName, year)

	if withWeekday {
		return fmt.Sprintf("%sที่ %s", ThaiWeekdayName(t), body)
	}

	return body
}

func joinDateTime(datePart string, t time.Time) string {
	if t.IsZero() || datePart == "" {
		return ""
	}
	return fmt.Sprintf("%s %s", datePart, t.Format("15:04"))
}

func withTimePtr(t *time.Time, fn func(time.Time) string) string {
	if t == nil {
		return ""
	}
	return fn(*t)
}
