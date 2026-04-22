package wela

import (
	"fmt"
	"time"
)

// ─────────────────────────────────────────────
// FORMAT: HUMAN READABLE DATE
// ─────────────────────────────────────────────

func FormatThaiDateLongBE(t time.Time) string {
	return formatThaiDate(t, false, true, false)
}

func FormatThaiDateLongCE(t time.Time) string {
	return formatThaiDate(t, false, false, false)
}

func FormatThaiDateShortBE(t time.Time) string {
	return formatThaiDate(t, true, true, false)
}

func FormatThaiDateShortCE(t time.Time) string {
	return formatThaiDate(t, true, false, false)
}

func FormatThaiDateLongWithDayBE(t time.Time) string {
	return formatThaiDate(t, false, true, true)
}

func FormatThaiDateLongWithDayCE(t time.Time) string {
	return formatThaiDate(t, false, false, true)
}

func FormatEnglishDateLong(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 January 2006")
}

func FormatEnglishDateShort(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006")
}

func FormatEnglishDateLongWithDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s, %s", EnglishWeekdayName(t), t.Format("2 January 2006"))
}

func FormatDateStyle(t time.Time, style DateFormat) string {
	if t.IsZero() {
		return ""
	}

	switch style {
	case DateThaiLongBE:
		return FormatThaiDateLongBE(t)
	case DateThaiLongCE:
		return FormatThaiDateLongCE(t)
	case DateThaiShortBE:
		return FormatThaiDateShortBE(t)
	case DateThaiShortCE:
		return FormatThaiDateShortCE(t)
	case DateEnglishLong:
		return FormatEnglishDateLong(t)
	case DateEnglishShort:
		return FormatEnglishDateShort(t)
	case DateThaiLongWithDayBE:
		return FormatThaiDateLongWithDayBE(t)
	case DateThaiLongWithDayCE:
		return FormatThaiDateLongWithDayCE(t)
	case DateEnglishLongWithDay:
		return FormatEnglishDateLongWithDay(t)
	default:
		return FormatDMYCE(t)
	}
}

// ─────────────────────────────────────────────
// FORMAT: HUMAN READABLE DATE + TIME
// ─────────────────────────────────────────────

func FormatThaiDateTimeLongBE(t time.Time) string {
	return joinDateTime(FormatThaiDateLongBE(t), t)
}

func FormatThaiDateTimeLongCE(t time.Time) string {
	return joinDateTime(FormatThaiDateLongCE(t), t)
}

func FormatThaiDateTimeShortBE(t time.Time) string {
	return joinDateTime(FormatThaiDateShortBE(t), t)
}

func FormatThaiDateTimeShortCE(t time.Time) string {
	return joinDateTime(FormatThaiDateShortCE(t), t)
}

func FormatThaiDateTimeLongWithDayBE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s เวลา %s", FormatThaiDateLongWithDayBE(t), t.Format("15:04"))
}

func FormatThaiDateTimeLongWithDayCE(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s เวลา %s", FormatThaiDateLongWithDayCE(t), t.Format("15:04"))
}

func FormatEnglishDateTimeLong(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 January 2006 15:04")
}

func FormatEnglishDateTimeShort(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006 15:04")
}

func FormatEnglishDateTimeLongWithDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s, %s", EnglishWeekdayName(t), t.Format("2 January 2006 15:04"))
}

func FormatEnglishDateThaiTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s เวลา %s", FormatEnglishDateLong(t), t.Format("15:04"))
}

func FormatDateTimeStyle(t time.Time, style DateTimeFormat) string {
	if t.IsZero() {
		return ""
	}

	switch style {
	case DateTimeThaiLongBE:
		return FormatThaiDateTimeLongBE(t)
	case DateTimeThaiLongCE:
		return FormatThaiDateTimeLongCE(t)
	case DateTimeThaiShortBE:
		return FormatThaiDateTimeShortBE(t)
	case DateTimeThaiShortCE:
		return FormatThaiDateTimeShortCE(t)
	case DateTimeEnglishLong:
		return FormatEnglishDateTimeLong(t)
	case DateTimeEnglishShort:
		return FormatEnglishDateTimeShort(t)
	case DateTimeThaiLongWithDayBE:
		return FormatThaiDateTimeLongWithDayBE(t)
	case DateTimeThaiLongWithDayCE:
		return FormatThaiDateTimeLongWithDayCE(t)
	case DateTimeEnglishLongWithDay:
		return FormatEnglishDateTimeLongWithDay(t)
	case DateTimeEnglishLongThaiTime:
		return FormatEnglishDateThaiTime(t)
	default:
		return FormatDMYTimeCE(t)
	}
}
