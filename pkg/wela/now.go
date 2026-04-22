package wela

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func NowBKK() time.Time {
	return time.Now().In(bangkokLoc)
}

func TodayBKK() time.Time {
	return StartOfDayBKK(time.Now())
}
