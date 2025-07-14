package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func RandInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// GenerateOrderNo creates a unique order number.
// e.g., "ORD20230715103050123456"
func GenerateOrderNo() string {
	now := time.Now()
	timestamp := now.Format("20060102150405")
	randomNum := rand.Intn(1000000)
	return fmt.Sprintf("ORD%s%06d", timestamp, randomNum)
}
