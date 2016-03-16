package gotrade

import (
	"time"
)

// 默认判断是否在开盘期间，支持自定义时间段 ([]string{"09:30:00", "11:30:00"}, []string{"13:00:00", "15:00:00"})
// 时间参数必须严格到 xx:xx:xx 否则会返回 false
func MarketIsOpening(durations ...[]string) bool {
	now := time.Now()
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}
	nowStr := now.Format("15:04:05")
	if len(durations) == 0 {
		if (nowStr >= "09:30:00" && nowStr <= "11:30:00") || (nowStr >= "13:00:00" && nowStr <= "15:00:00") {
			return true
		}
		return false
	}
	for _, duration := range durations {
		if len(duration) != 2 {
			continue
		}
		var start, end time.Time
		var err error
		start, err = time.Parse("15:04:05", duration[0])
		if err != nil {
			continue
		}
		end, err = time.Parse("15:04:05", duration[1])
		if err != nil {
			continue
		}
		if nowStr >= start.Format("15:04:05") && nowStr <= end.Format("15:04:05") {
			return true
		}
	}
	return false
}
