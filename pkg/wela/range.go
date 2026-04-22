package wela

import "time"

func StartOfDayBKK(t time.Time) time.Time {
	bkk := t.In(bangkokLoc)
	return time.Date(bkk.Year(), bkk.Month(), bkk.Day(), 0, 0, 0, 0, bangkokLoc)
}

func StartOfNextDayBKK(t time.Time) time.Time {
	return StartOfDayBKK(t).AddDate(0, 0, 1)
}

func EndOfDayBKK(t time.Time) time.Time {
	return StartOfNextDayBKK(t).Add(-time.Nanosecond)
}

func StartOfMonthBKK(t time.Time) time.Time {
	bkk := t.In(bangkokLoc)
	return time.Date(bkk.Year(), bkk.Month(), 1, 0, 0, 0, 0, bangkokLoc)
}

func StartOfNextMonthBKK(t time.Time) time.Time {
	return StartOfMonthBKK(t).AddDate(0, 1, 0)
}

func EndOfMonthBKK(t time.Time) time.Time {
	return StartOfNextMonthBKK(t).Add(-time.Nanosecond)
}

func SameDayBKK(a, b time.Time) bool {
	a = a.In(bangkokLoc)
	b = b.In(bangkokLoc)
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// DayRangeBKK คืนขอบเขตของวันสำหรับ database query แบบ exclusive end
// ใช้คู่กับ WHERE created_at >= ? AND created_at < ?
func DayRangeBKK(t time.Time) (start time.Time, endExclusive time.Time) {
	start = StartOfDayBKK(t)
	endExclusive = StartOfNextDayBKK(t)
	return
}

// MonthRangeBKK คืนขอบเขตของเดือนสำหรับ database query แบบ exclusive end
// ใช้คู่กับ WHERE created_at >= ? AND created_at < ?
func MonthRangeBKK(t time.Time) (start time.Time, endExclusive time.Time) {
	start = StartOfMonthBKK(t)
	endExclusive = StartOfNextMonthBKK(t)
	return
}
