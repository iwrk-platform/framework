package frontend

import (
	"strconv"
	"time"
)

func IntToString(i int) string {
	return strconv.Itoa(i)
}

func UintToString(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Uint64ToString(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func Int32ToString(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func Uint32ToString(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

func Float64ToString(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func Float32ToString(n float32) string {
	return strconv.FormatFloat(float64(n), 'f', -1, 32)
}

func TimeStringToInt64(s string) int64 {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0
	}
	return t.AddDate(1970, 0, 0).UTC().Unix()
}

// TimeInt64ToString HH:MM formated
func TimeInt64ToString(i int64) string {
	return time.Unix(i, 0).Format("15:04")
}

// TimeInt64ToMSString MM:SS formated
func TimeInt64ToMSString(i int64) string {
	return time.Unix(i, 0).Format("04:05")
}

func DateStringToInt64(s string) int64 {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func DateInt64ToString(i int64, format string) string {
	return time.Unix(i, 0).Format(format)
}

func DateInt64ToHtmlString(i int64) string {
	return time.Unix(i, 0).Format("2006-01-02")
}
