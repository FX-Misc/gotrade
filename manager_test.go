package gotrade

import (
	"gotrade"
	// "log"
	"testing"
	// "time"
)

func Test_Run(t *testing.T) {
	manager := gotrade.NewManager()
	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	manager.Start()
	// 	log.Println("pause")
	// }()
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	manager.Pause()
	// 	log.Println("pause")
	// }()
	manager.Listen()
}
