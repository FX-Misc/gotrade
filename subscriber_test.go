package gotrade

import (
	// "log"
	"testing"
)

var sbr *Subscriber

func init() {
	sbr = NewSubscriber("config/subscriber.yaml")
}

func TestRun(t *testing.T) {
	codeList := []string{"150168", "150167"}
	quoChan := sbr.Subscribe("test", codeList)
	ticketChan := sbr.SubscribeTicket("test", codeList)

	go sbr.Run()
	go func() {
		for _ = range quoChan {
		}
	}()

	go func() {
		for _ = range ticketChan {
		}
	}()

	select {}
}
