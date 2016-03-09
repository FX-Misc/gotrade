package gotrade

import (
	"time"
)

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
