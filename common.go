package gotrade

import (
	"time"
)

func TimeWait(timeString string) {
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
