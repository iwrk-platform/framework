package frontend

import (
	"strconv"
	"time"
)

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

func TimeInt64ToString(i int64) string {
	return time.Unix(i, 0).Format("15:04")
}
