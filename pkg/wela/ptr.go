package wela

import "time"

// Pointer Wrappers — ใช้เมื่อ field เป็น nullable จาก database หรือ JSON
// ถ้า pointer เป็น nil จะคืน ""

func FormatDMYCEPtr(t *time.Time) string           { return withTimePtr(t, FormatDMYCE) }
func FormatDMYTimeCEPtr(t *time.Time) string       { return withTimePtr(t, FormatDMYTimeCE) }
func FormatDMYBEPtr(t *time.Time) string           { return withTimePtr(t, FormatDMYBE) }
func FormatDMYTimeBEPtr(t *time.Time) string       { return withTimePtr(t, FormatDMYTimeBE) }
func FormatISODatePtr(t *time.Time) string         { return withTimePtr(t, FormatISODate) }
func FormatISODateTimePtr(t *time.Time) string     { return withTimePtr(t, FormatISODateTime) }
func FormatRFC3339Ptr(t *time.Time) string         { return withTimePtr(t, FormatRFC3339) }
func FormatThaiDateLongBEPtr(t *time.Time) string  { return withTimePtr(t, FormatThaiDateLongBE) }
func FormatThaiDateLongCEPtr(t *time.Time) string  { return withTimePtr(t, FormatThaiDateLongCE) }
func FormatThaiDateShortBEPtr(t *time.Time) string { return withTimePtr(t, FormatThaiDateShortBE) }
func FormatThaiDateShortCEPtr(t *time.Time) string { return withTimePtr(t, FormatThaiDateShortCE) }
func FormatThaiDateLongWithDayBEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateLongWithDayBE)
}
func FormatThaiDateLongWithDayCEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateLongWithDayCE)
}
func FormatEnglishDateLongPtr(t *time.Time) string  { return withTimePtr(t, FormatEnglishDateLong) }
func FormatEnglishDateShortPtr(t *time.Time) string { return withTimePtr(t, FormatEnglishDateShort) }
func FormatEnglishDateLongWithDayPtr(t *time.Time) string {
	return withTimePtr(t, FormatEnglishDateLongWithDay)
}
func FormatThaiDateTimeLongBEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeLongBE)
}
func FormatThaiDateTimeLongCEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeLongCE)
}
func FormatThaiDateTimeShortBEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeShortBE)
}
func FormatThaiDateTimeShortCEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeShortCE)
}
func FormatThaiDateTimeLongWithDayBEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeLongWithDayBE)
}
func FormatThaiDateTimeLongWithDayCEPtr(t *time.Time) string {
	return withTimePtr(t, FormatThaiDateTimeLongWithDayCE)
}
func FormatEnglishDateTimeLongPtr(t *time.Time) string {
	return withTimePtr(t, FormatEnglishDateTimeLong)
}
func FormatEnglishDateTimeShortPtr(t *time.Time) string {
	return withTimePtr(t, FormatEnglishDateTimeShort)
}
func FormatEnglishDateTimeLongWithDayPtr(t *time.Time) string {
	return withTimePtr(t, FormatEnglishDateTimeLongWithDay)
}
func FormatEnglishDateThaiTimePtr(t *time.Time) string {
	return withTimePtr(t, FormatEnglishDateThaiTime)
}
func ReadDateThaiSimplePtr(t *time.Time) string     { return withTimePtr(t, ReadDateThaiSimple) }
func ReadDateThaiFullPtr(t *time.Time) string       { return withTimePtr(t, ReadDateThaiFull) }
func ReadDateTimeThaiSimplePtr(t *time.Time) string { return withTimePtr(t, ReadDateTimeThaiSimple) }
func ReadDateTimeThaiFullPtr(t *time.Time) string   { return withTimePtr(t, ReadDateTimeThaiFull) }
