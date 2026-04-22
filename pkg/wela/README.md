# wela

Package `wela` (เวลา) รวม Utility สำหรับจัดการ `time.Time` ในระบบที่ใช้ Timezone Bangkok (`Asia/Bangkok`) และปีพุทธศักราช (พ.ศ.)

> **Note:** Package นี้ Embed ข้อมูล `time/tzdata` ไว้ใน Binary แล้ว ไม่จำเป็นต้องติดตั้งแพ็กเกจ `tzdata` เพิ่มเติมใน Docker image เช่น Alpine หรือ Scratch

> **⚠️ Codebase Convention:** `wela` คือ standard สำหรับจัดการเวลาใน codebase นี้ทั้งหมด
> ห้ามใช้ `time.Now()`, `time.Parse()`, หรือ layout string โดยตรง — ให้ใช้ฟังก์ชันของ `wela` แทนเสมอ

---

## การติดตั้งและนำไปใช้งาน

```go
import "vexentra-api/pkg/wela"
```

---

## แนวคิดของ package นี้

`wela` ถูกออกแบบมาสำหรับงาน backend / API / database ที่ต้องจัดการเรื่อง:

- เวลาในเขต `Asia/Bangkok`
- ปี ค.ศ. และ พ.ศ.
- การ parse วันที่จากหลายรูปแบบ
- การ format วันที่สำหรับแสดงผลทั้ง machine-readable และ human-readable
- การสร้างขอบเขตเวลาแบบปลอดภัยสำหรับ query ฐานข้อมูล

---

## ภาพรวม API

| หมวด                  | Functions                                                                                                                                                                                                                                                                                                                                 |
| :-------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Now**               | `NowUTC`, `NowBKK`, `TodayBKK`                                                                                                                                                                                                                                                                                                            |
| **Timezone**          | `InBKK`, `InUTC`                                                                                                                                                                                                                                                                                                                          |
| **Year**              | `YearCEtoBE`, `YearBEtoCE`                                                                                                                                                                                                                                                                                                                |
| **Parse**             | `ParseISODateBKK`, `ParseISODateTimeBKK`, `ParseISODateUTC`, `ParseISODateTimeUTC`, `ParseRFC3339Any`, `ParseRFC3339BKK`, `ParseThaiDateBKK`, `ParseThaiDateTimeBKK`, `ParseDMYBKK`, `ParseDateBKKAuto`                                                                                                                                   |
| **Basic Format**      | `FormatDMYCE`, `FormatDMYTimeCE`, `FormatDMYBE`, `FormatDMYTimeBE`, `FormatISODate`, `FormatISODateTime`, `FormatRFC3339`, `FormatCustom`, `FormatInBKK`                                                                                                                                                                                  |
| **Readable Date**     | `FormatThaiDateLongBE`, `FormatThaiDateLongCE`, `FormatThaiDateShortBE`, `FormatThaiDateShortCE`, `FormatThaiDateLongWithDayBE`, `FormatThaiDateLongWithDayCE`, `FormatEnglishDateLong`, `FormatEnglishDateShort`, `FormatEnglishDateLongWithDay`, `FormatDateStyle`                                                                      |
| **Readable DateTime** | `FormatThaiDateTimeLongBE`, `FormatThaiDateTimeLongCE`, `FormatThaiDateTimeShortBE`, `FormatThaiDateTimeShortCE`, `FormatThaiDateTimeLongWithDayBE`, `FormatThaiDateTimeLongWithDayCE`, `FormatEnglishDateTimeLong`, `FormatEnglishDateTimeShort`, `FormatEnglishDateTimeLongWithDay`, `FormatEnglishDateThaiTime`, `FormatDateTimeStyle` |
| **Weekday**           | `ThaiWeekdayName`, `EnglishWeekdayName`                                                                                                                                                                                                                                                                                                   |
| **Reading**           | `ReadDateThaiSimple`, `ReadDateThaiFull`, `ReadDateTimeThaiSimple`, `ReadDateTimeThaiFull`, `ThaiNumberToWords`                                                                                                                                                                                                                           |
| **Pointer Wrappers**  | ฟังก์ชันกลุ่ม `...Ptr` สำหรับ format / read                                                                                                                                                                                                                                                                                               |
| **Utils**             | `StartOfDayBKK`, `StartOfNextDayBKK`, `EndOfDayBKK`, `StartOfMonthBKK`, `StartOfNextMonthBKK`, `EndOfMonthBKK`, `SameDayBKK`, `DayRangeBKK`, `MonthRangeBKK`                                                                                                                                                                              |

---

## 1. การดึงเวลาปัจจุบัน (Now)

```go
wela.NowUTC()   // 2026-04-22 03:00:00 +0000 UTC
wela.NowBKK()   // 2026-04-22 10:00:00 +0700 ICT
wela.TodayBKK() // 2026-04-22 00:00:00 +0700 ICT
```

---

## 2. การจัดการ Timezone

เปลี่ยน Timezone สำหรับการแสดงผล โดยที่ **instant เดิมไม่เปลี่ยน**

```go
utc := wela.NowUTC()
bkk := wela.InBKK(utc)

bkk2 := wela.NowBKK()
utc2 := wela.InUTC(bkk2)
```

---

## 3. การแปลงปี ค.ศ. ↔ พ.ศ.

```go
wela.YearCEtoBE(2026) // 2569
wela.YearBEtoCE(2569) // 2026
```

---

## 4. การ Parse (String → time.Time)

ฟังก์ชันกลุ่ม Parse คืนค่า `(time.Time, error)` ตามมาตรฐานของ Go

- ถ้า input เป็น `""` จะคืน `time.Time{}` และ `nil`
- ถ้า parse ไม่ได้ หรือวันที่ไม่มีจริง จะคืน `time.Time{}` และ `error`

### 4.1 Parse ISO

```go
t1, err := wela.ParseISODateBKK("2026-01-16")
// 2026-01-16 00:00:00 +0700 ICT

t2, err := wela.ParseISODateTimeUTC("2026-01-16 14:30:00")
```

### 4.2 Parse Thai Date

รองรับรูปแบบ `DD/MM/พ.ศ.`

```go
t, err := wela.ParseThaiDateBKK("16/01/2569")
// 2026-01-16 00:00:00 +0700 ICT
```

รองรับ `DD/MM/พ.ศ. HH:MM` และ `DD/MM/พ.ศ. HH:MM:SS`

```go
t, err := wela.ParseThaiDateTimeBKK("16/01/2569 14:30")
t2, err := wela.ParseThaiDateTimeBKK("16/01/2569 14:30:45")
```

> **Strict Calendar Validation:** ถ้าวันที่ไม่มีจริง เช่น `"31/04/2569"` จะคืน error ทันที

### 4.3 Parse RFC3339

```go
t, err := wela.ParseRFC3339Any("2026-01-16T14:30:00+07:00")
t2, err := wela.ParseRFC3339BKK("2026-01-16T14:30:00Z")
```

- `ParseRFC3339Any` คืนค่า instant ตาม timezone ที่ input ส่งมา
- `ParseRFC3339BKK` จะ convert ผลลัพธ์ให้แสดงใน timezone Bangkok

### 4.4 Parse Auto

ใช้เมื่อ input จาก UI / API อาจมีหลายรูปแบบ

```go
t1, _ := wela.ParseDateBKKAuto("2026-01-16")
t2, _ := wela.ParseDateBKKAuto("2026-01-16 14:30:00")
t3, _ := wela.ParseDateBKKAuto("16/01/2026")
t4, _ := wela.ParseDateBKKAuto("16/01/2569")
t5, _ := wela.ParseDateBKKAuto("16/01/2569 14:30")
t6, _ := wela.ParseDateBKKAuto("2026-01-16T14:30:00+07:00")
```

เหมาะกับ endpoint ที่รับ date string หลายรูปแบบจาก frontend

---

## 5. การ Format แบบพื้นฐาน (time.Time → string)

ทุกฟังก์ชันเป็น **Zero-safe** — ถ้า `t.IsZero()` จะคืน `""`

```go
t := time.Date(2026, 1, 16, 14, 30, 0, 0, time.UTC)

wela.FormatDMYCE(t)      // "16/01/2026"
wela.FormatDMYTimeCE(t)  // "16/01/2026 14:30"

wela.FormatDMYBE(t)      // "16/01/2569"
wela.FormatDMYTimeBE(t)  // "16/01/2569 14:30"

wela.FormatISODate(t)      // "2026-01-16"
wela.FormatISODateTime(t)  // "2026-01-16 14:30:00"
wela.FormatRFC3339(t)      // "2026-01-16T14:30:00Z"

wela.FormatCustom(t, "Jan 02, 2006")        // "Jan 16, 2026"
wela.FormatInBKK(t, "02/01/2006 15:04:05")  // "16/01/2026 21:30:00"
```

---

## 6. การ Format แบบอ่านง่าย (Readable Date)

### 6.1 ภาษาไทย

```go
loc, _ := time.LoadLocation("Asia/Bangkok")
t := time.Date(2026, 1, 16, 14, 30, 0, 0, loc)

wela.FormatThaiDateLongBE(t)        // "16 มกราคม 2569"
wela.FormatThaiDateLongCE(t)        // "16 มกราคม 2026"
wela.FormatThaiDateShortBE(t)       // "16 ม.ค. 2569"
wela.FormatThaiDateShortCE(t)       // "16 ม.ค. 2026"

wela.FormatThaiDateLongWithDayBE(t) // "วันศุกร์ที่ 16 มกราคม 2569"
wela.FormatThaiDateLongWithDayCE(t) // "วันศุกร์ที่ 16 มกราคม 2026"
```

### 6.2 ภาษาอังกฤษ

```go
wela.FormatEnglishDateLong(t)         // "16 January 2026"
wela.FormatEnglishDateShort(t)        // "16 Jan 2026"
wela.FormatEnglishDateLongWithDay(t)  // "Friday, 16 January 2026"
```

### 6.3 ใช้ style กลาง

```go
wela.FormatDateStyle(t, wela.DateThaiLongBE)
wela.FormatDateStyle(t, wela.DateThaiShortBE)
wela.FormatDateStyle(t, wela.DateEnglishLong)
wela.FormatDateStyle(t, wela.DateEnglishLongWithDay)
```

เหมาะกับกรณีที่ style ถูกกำหนดจาก config หรือ enum จาก service layer

---

## 7. การ Format วันที่ + เวลา (Readable DateTime)

### 7.1 ภาษาไทย

```go
wela.FormatThaiDateTimeLongBE(t)         // "16 มกราคม 2569 14:30"
wela.FormatThaiDateTimeShortBE(t)        // "16 ม.ค. 2569 14:30"
wela.FormatThaiDateTimeLongWithDayBE(t)  // "วันศุกร์ที่ 16 มกราคม 2569 เวลา 14:30"
```

### 7.2 ภาษาอังกฤษ

```go
wela.FormatEnglishDateTimeLong(t)          // "16 January 2026 14:30"
wela.FormatEnglishDateTimeShort(t)         // "16 Jan 2026 14:30"
wela.FormatEnglishDateTimeLongWithDay(t)   // "Friday, 16 January 2026 14:30"
wela.FormatEnglishDateThaiTime(t)          // "16 January 2026 เวลา 14:30"
```

### 7.3 ใช้ style กลาง

```go
wela.FormatDateTimeStyle(t, wela.DateTimeThaiLongBE)
wela.FormatDateTimeStyle(t, wela.DateTimeThaiLongWithDayBE)
wela.FormatDateTimeStyle(t, wela.DateTimeEnglishLong)
wela.FormatDateTimeStyle(t, wela.DateTimeEnglishLongThaiTime)
```

---

## 8. ชื่อวันในสัปดาห์

```go
wela.ThaiWeekdayName(t)    // "วันศุกร์"
wela.EnglishWeekdayName(t) // "Friday"
```

---

## 9. คำอ่านภาษาไทย

เหมาะกับงาน TTS, voice prompt, chatbot, หรือ UI ที่ต้องการแสดงคำอ่าน

### 9.1 แบบง่าย

```go
wela.ReadDateThaiSimple(t)
// "วันที่ 16 เดือนมกราคม ปี 2569"

wela.ReadDateTimeThaiSimple(t)
// "วันที่ 16 เดือนมกราคม ปี 2569 เวลา 14 นาฬิกา 30 นาที"
```

### 9.2 แบบเต็ม

```go
wela.ReadDateThaiFull(t)
// "สิบหก มกราคม สองพันห้าร้อยหกสิบเก้า"

wela.ReadDateTimeThaiFull(t)
// "สิบหก มกราคม สองพันห้าร้อยหกสิบเก้า เวลา สิบสี่ นาฬิกา สามสิบ นาที"
```

### 9.3 แปลงตัวเลขเป็นคำไทย

```go
wela.ThaiNumberToWords(16)    // "สิบหก"
wela.ThaiNumberToWords(21)    // "ยี่สิบเอ็ด"
wela.ThaiNumberToWords(2569)  // "สองพันห้าร้อยหกสิบเก้า"
```

---

## 10. Pointer Wrappers (`*time.Time`)

ใช้เมื่อ field เป็น nullable จาก database หรือ JSON

```go
var deletedAt *time.Time
wela.FormatISODatePtr(deletedAt) // ""

now := time.Now()
wela.FormatThaiDateLongBEPtr(&now)
wela.FormatEnglishDateTimeLongPtr(&now)
wela.ReadDateThaiFullPtr(&now)
```

หลักการคือ:

- ถ้า pointer เป็น `nil` → คืน `""`
- ถ้ามีค่า → เรียก formatter/read function ตัวจริงให้

---

## 11. Utils สำหรับวัน/เดือน และ database query

### 11.1 ขอบเขตวัน

```go
t := time.Date(2026, 1, 16, 10, 0, 0, 0, time.UTC)

wela.StartOfDayBKK(t)
wela.StartOfNextDayBKK(t)
wela.EndOfDayBKK(t)
```

### 11.2 ขอบเขตเดือน

```go
wela.StartOfMonthBKK(t)
wela.StartOfNextMonthBKK(t)
wela.EndOfMonthBKK(t)
```

### 11.3 ตรวจว่าเป็นวันเดียวกันหรือไม่

```go
wela.SameDayBKK(a, b)
```

---

## 12. Best Practice สำหรับ Database Query

แม้จะมี `EndOfDayBKK` และ `EndOfMonthBKK` ให้ใช้ แต่สำหรับ database จริง  
**แนะนำให้ใช้ช่วงแบบ exclusive end (`>=` และ `<`) เสมอ**

### 12.1 Query รายวัน

```go
start, endExclusive := wela.DayRangeBKK(time.Now())

db.Where("created_at >= ? AND created_at < ?", start.UTC(), endExclusive.UTC())
```

### 12.2 Query รายเดือน

```go
start, endExclusive := wela.MonthRangeBKK(time.Now())

db.Where("created_at >= ? AND created_at < ?", start.UTC(), endExclusive.UTC())
```

ข้อดี:

- ปลอดภัยกว่า `BETWEEN`
- ไม่เสี่ยง precision rounding
- ใช้งานได้เสถียรกับ PostgreSQL / MySQL / อื่น ๆ

---

## 13. โครงสร้างไฟล์

```
pkg/wela/
├── wela.go          # constants, types, vars, timezone/year helpers
├── now.go           # NowUTC, NowBKK, TodayBKK
├── parse.go         # Parse functions
├── format_basic.go  # FormatDMY*, FormatISO*, FormatRFC3339, weekday
├── format_human.go  # FormatThai*, FormatEnglish*, FormatDateStyle, FormatDateTimeStyle
├── read.go          # ReadDate*, ThaiNumberToWords
├── range.go         # StartOf*, EndOf*, SameDayBKK, DayRangeBKK, MonthRangeBKK
├── ptr.go           # Pointer wrappers
└── internal.go      # helper functions (unexported)
```

---

## 14. Key Behaviors

### Empty String Tolerance

ฟังก์ชัน parse รองรับ string ว่างโดยไม่ panic และไม่คืน error

```go
t, err := wela.ParseISODateBKK("")
// t == time.Time{}, err == nil
```

### Anti-Silent Normalization

ป้องกันพฤติกรรมที่ `time.Date` แอบ normalize วันที่ผิด เช่น:

- `31/04/2569` ไม่ถูกแปลงเป็น `01/05/2026` แต่จะคืน error ทันที

### Zero-safe Formatting

ฟังก์ชัน format / read ทั้งหมดจะคืน `""` ถ้า input เป็น zero time

---

## 15. ตัวอย่างใช้งานรวม

```go
t, err := wela.ParseDateBKKAuto("16/01/2569 14:30")
if err != nil {
    panic(err)
}

fmt.Println(wela.FormatThaiDateLongBE(t))         // 16 มกราคม 2569
fmt.Println(wela.FormatEnglishDateLong(t))        // 16 January 2026
fmt.Println(wela.FormatThaiDateLongWithDayBE(t))  // วันศุกร์ที่ 16 มกราคม 2569
fmt.Println(wela.FormatEnglishDateThaiTime(t))    // 16 January 2026 เวลา 14:30

fmt.Println(wela.ReadDateThaiSimple(t))           // วันที่ 16 เดือนมกราคม ปี 2569
fmt.Println(wela.ReadDateThaiFull(t))             // สิบหก มกราคม สองพันห้าร้อยหกสิบเก้า

start, endExclusive := wela.MonthRangeBKK(t)
fmt.Println(start, endExclusive)
```
