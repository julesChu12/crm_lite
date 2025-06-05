package timeutil

import (
	"time"
)

const (
	DefaultFormat = "2006-01-02 15:04:05"
	DateFormat    = "2006-01-02"
)

// FormatTime 将 time.Time 格式化为标准字符串 (YYYY-MM-DD HH:MM:SS)
func FormatTime(t time.Time) string {
	return t.Format(DefaultFormat)
}

// FormatDate 将 time.Time 格式化为日期字符串 (YYYY-MM-DD)
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// ParseTime 将标准时间字符串解析为 time.Time
func ParseTime(s string) (time.Time, error) {
	return time.ParseInLocation(DefaultFormat, s, time.Local) // 使用本地时区
}

// ParseDate 将日期字符串解析为 time.Time
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, s, time.Local)
}

// GetCurrentTime 获取当前时间字符串
func GetCurrentTime() string {
	return FormatTime(time.Now())
}

// GetCurrentDate 获取当前日期字符串
func GetCurrentDate() string {
	return FormatDate(time.Now())
}
