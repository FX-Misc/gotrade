package gotrade

import (
	"testing"
)

func TestSinaSubscriber(t *testing.T) {
	sbrSina := NewSinaSubscriber("config/subscriber.yaml")
	codeList := []string{"150168", "150167"}
	quoChan := sbrSina.Subscribe("test", codeList)
	ticketChan := sbrSina.SubscribeTicket("test", codeList)

	go sbrSina.Run()
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
