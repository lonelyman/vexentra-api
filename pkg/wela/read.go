package wela

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ReadDateThaiSimple คืนคำอ่านวันที่แบบง่าย เช่น "วันที่ 16 เดือนมกราคม ปี 2569"
func ReadDateThaiSimple(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("วันที่ %d เดือน%s ปี %d",
		t.Day(),
		thaiMonthsFull[int(t.Month())],
		YearCEtoBE(t.Year()),
	)
}

// ReadDateThaiFull คืนคำอ่านวันที่แบบเต็ม เช่น "สิบหก มกราคม สองพันห้าร้อยหกสิบเก้า"
func ReadDateThaiFull(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s %s %s",
		ThaiNumberToWords(t.Day()),
		thaiMonthsFull[int(t.Month())],
		ThaiNumberToWords(YearCEtoBE(t.Year())),
	)
}

// ReadDateTimeThaiSimple คืนคำอ่านวันที่+เวลาแบบง่าย
func ReadDateTimeThaiSimple(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("วันที่ %d เดือน%s ปี %d เวลา %d นาฬิกา %d นาที",
		t.Day(),
		thaiMonthsFull[int(t.Month())],
		YearCEtoBE(t.Year()),
		t.Hour(),
		t.Minute(),
	)
}

// ReadDateTimeThaiFull คืนคำอ่านวันที่+เวลาแบบเต็ม
func ReadDateTimeThaiFull(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s %s %s เวลา %s นาฬิกา %s นาที",
		ThaiNumberToWords(t.Day()),
		thaiMonthsFull[int(t.Month())],
		ThaiNumberToWords(YearCEtoBE(t.Year())),
		ThaiNumberToWords(t.Hour()),
		ThaiNumberToWords(t.Minute()),
	)
}

// ThaiNumberToWords แปลงตัวเลขเป็นคำอ่านภาษาไทย
func ThaiNumberToWords(n int) string {
	if n == 0 {
		return "ศูนย์"
	}
	if n < 0 {
		return "ลบ" + ThaiNumberToWords(-n)
	}

	digits := []string{"", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}
	positions := []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน"}

	if n >= 1000000 {
		left := n / 1000000
		right := n % 1000000
		if right == 0 {
			return ThaiNumberToWords(left) + "ล้าน"
		}
		return ThaiNumberToWords(left) + "ล้าน" + ThaiNumberToWords(right)
	}

	s := strconv.Itoa(n)
	var b strings.Builder
	length := len(s)

	for i, r := range s {
		d := int(r - '0')
		pos := length - i - 1

		if d == 0 {
			continue
		}

		switch pos {
		case 0:
			if d == 1 && length > 1 {
				b.WriteString("เอ็ด")
			} else {
				b.WriteString(digits[d])
			}
		case 1:
			if d == 1 {
				b.WriteString("สิบ")
			} else if d == 2 {
				b.WriteString("ยี่สิบ")
			} else {
				b.WriteString(digits[d])
				b.WriteString("สิบ")
			}
		default:
			b.WriteString(digits[d])
			if pos < len(positions) {
				b.WriteString(positions[pos])
			}
		}
	}

	return b.String()
}
