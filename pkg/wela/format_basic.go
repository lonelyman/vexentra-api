package wela

import (
	"fmt"
	"time"
)

func FormatDMYCE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutDMY)
}

func FormatDMYTimeCE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutDMYTime)
}

func FormatDMYBE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), int(t.Month()), YearCEtoBE(t.Year()))
}

func FormatDMYTimeBE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%d %s", t.Day(), int(t.Month()), YearCEtoBE(t.Year()), t.Format("15:04"))
}

func FormatISODate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutISODate)
}

func FormatISODateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutISODateTime)
}

func FormatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func FormatCustom(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

func FormatInBKK(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.In(bangkokLoc).Format(layout)
}

func ThaiWeekdayName(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return thaiWeekdaysFull[int(t.Weekday())]
}

func EnglishWeekdayName(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return englishWeekdaysFull[int(t.Weekday())]
}
