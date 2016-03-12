package gotrade

import (
	"time"
)

// 是否在开盘期间 (周一到周五 9:30-11:30 13:00-15:00)
func MarketIsOpening() bool {
	now := time.Now()
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}
	if now.Hour() < 9 {
		return false
	}
	if now.Hour() == 9 && now.Minute() < 30 {
		return false
	}
	if now.Hour() == 11 && now.Minute() >= 30 {
		return false
	}
	if now.Hour() == 12 {
		return false
	}
	if now.Hour() == 15 && now.Minute() >= 1 {
		return false
	}
	if now.Hour() > 15 {
		return false
	}
	return true
}

// @deprecated
func WaitUntilLater(timeString string) {
	for {
		// check time
		nowString := time.Now().Format("15:04:05")
		if nowString <= timeString {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

// @deprecated
func WaitUntilEalier(timeString string) {
	for {
		// check time
		nowString := time.Now().Format("15:04:05")
		if nowString >= timeString {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}
