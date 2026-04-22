// Package wela (เวลา) รวม Utility สำหรับจัดการ time.Time
// ในระบบที่ใช้ Timezone Bangkok (Asia/Bangkok) และปีพุทธศักราช (พ.ศ.)
// Package นี้ Embed ข้อมูล time/tzdata ไว้ใน Binary แล้ว
// ไม่จำเป็นต้องติดตั้งแพ็กเกจ tzdata เพิ่มเติมใน Docker image
package wela

import (
	"time"
	_ "time/tzdata"
)

const (
	LayoutISODate     = "2006-01-02"
	LayoutISODateTime = "2006-01-02 15:04:05"
	LayoutDMY         = "02/01/2006"
	LayoutDMYTime     = "02/01/2006 15:04"
	LayoutRFC3339     = time.RFC3339
)

type DateFormat string
type DateTimeFormat string

const (
	// Date only
	DateThaiLongBE         DateFormat = "thai_long_be"
	DateThaiLongCE         DateFormat = "thai_long_ce"
	DateThaiShortBE        DateFormat = "thai_short_be"
	DateThaiShortCE        DateFormat = "thai_short_ce"
	DateEnglishLong        DateFormat = "english_long"
	DateEnglishShort       DateFormat = "english_short"
	DateThaiLongWithDayBE  DateFormat = "thai_long_with_day_be"
	DateThaiLongWithDayCE  DateFormat = "thai_long_with_day_ce"
	DateEnglishLongWithDay DateFormat = "english_long_with_day"

	// Date + Time
	DateTimeThaiLongBE          DateTimeFormat = "thai_long_be"
	DateTimeThaiLongCE          DateTimeFormat = "thai_long_ce"
	DateTimeThaiShortBE         DateTimeFormat = "thai_short_be"
	DateTimeThaiShortCE         DateTimeFormat = "thai_short_ce"
	DateTimeEnglishLong         DateTimeFormat = "english_long"
	DateTimeEnglishShort        DateTimeFormat = "english_short"
	DateTimeThaiLongWithDayBE   DateTimeFormat = "thai_long_with_day_be"
	DateTimeThaiLongWithDayCE   DateTimeFormat = "thai_long_with_day_ce"
	DateTimeEnglishLongWithDay  DateTimeFormat = "english_long_with_day"
	DateTimeEnglishLongThaiTime DateTimeFormat = "english_long_thai_time"
)

// bangkokLoc โหลด timezone Bangkok ครั้งเดียว (package-level)
var bangkokLoc *time.Location

func init() {
	var err error
	bangkokLoc, err = time.LoadLocation("Asia/Bangkok")
	if err != nil {
		bangkokLoc = time.FixedZone("ICT", 7*60*60)
	}
}

var thaiMonthsFull = []string{
	"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}

var thaiMonthsShort = []string{
	"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
	"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
}

var thaiWeekdaysFull = []string{
	"วันอาทิตย์", "วันจันทร์", "วันอังคาร", "วันพุธ", "วันพฤหัสบดี", "วันศุกร์", "วันเสาร์",
}

var englishWeekdaysFull = []string{
	"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
}

// InBKK แปลง time.Time ให้แสดงใน timezone Bangkok
func InBKK(t time.Time) time.Time {
	return t.In(bangkokLoc)
}

// InUTC แปลง time.Time ให้แสดงใน timezone UTC
func InUTC(t time.Time) time.Time {
	return t.UTC()
}

// YearCEtoBE แปลงปี ค.ศ. เป็น พ.ศ.
func YearCEtoBE(year int) int {
	return year + 543
}

// YearBEtoCE แปลงปี พ.ศ. เป็น ค.ศ.
func YearBEtoCE(year int) int {
	return year - 543
}
