package utils

import "time"

func Now() time.Time {
	return time.Now()
}

// DefaultTimeFormat 是一个标准的时间格式化模板
const DefaultTimeFormat = "2006-01-02 15:04:05"

// FormatTime 将 time.Time 格式化为标准字符串
// 如果时间是零值，则返回空字符串
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DefaultTimeFormat)
}
