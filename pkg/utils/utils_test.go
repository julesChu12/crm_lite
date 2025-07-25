package utils

import (
	"context"
	"testing"
	"time"
)

func TestRandInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := RandInt(5, 10)
		if v < 5 || v > 10 {
			t.Fatalf("RandInt out of range: %d", v)
		}
	}
}

func TestGenerateOrderNo(t *testing.T) {
	n1 := GenerateOrderNo()
	time.Sleep(time.Millisecond) // 确保时间戳不同
	n2 := GenerateOrderNo()

	if len(n1) < len("ORD")+14+6 { // 前缀+时间戳+随机数
		t.Errorf("order number too short: %s", n1)
	}
	if n1 == n2 {
		t.Errorf("order numbers should be unique")
	}
	if n1[:3] != "ORD" {
		t.Errorf("order number prefix incorrect: %s", n1)
	}
}

func TestTimeUtils(t *testing.T) {
	var zero time.Time
	if s := FormatTime(zero); s != "" {
		t.Errorf("expected empty string for zero time, got %s", s)
	}

	sample := time.Date(2023, 7, 15, 10, 30, 50, 0, time.UTC)
	expected := "2023-07-15 10:30:50"
	if s := FormatTime(sample); s != expected {
		t.Errorf("expected %s, got %s", expected, s)
	}
}

func TestContextUtils(t *testing.T) {
	ctx := WithUser(context.Background(), "u1", "tester")
	uid, ok := GetUserID(ctx)
	if !ok || uid != "u1" {
		t.Errorf("GetUserID failed: %s %v", uid, ok)
	}
	name, ok := GetUsername(ctx)
	if !ok || name != "tester" {
		t.Errorf("GetUsername failed: %s %v", name, ok)
	}
}
