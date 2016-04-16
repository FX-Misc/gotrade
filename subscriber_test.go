package gotrade

import (
	"log"
	"testing"
)

var sbr *Subscriber

func init() {
	sbr = NewSubscriber("config/subscriber.yaml")
}

func TestRun(t *testing.T) {
	codeList := []string{"150168", "150167", "150250", "i000300", "150250", "150250"}
	quoChan := sbr.Subscribe("test", codeList)
	ticketChan := sbr.SubscribeTicket("test", codeList)

	go sbr.Run()
	go func() {
		for quo := range quoChan {
			log.Println(quo)
		}
	}()

	go func() {
		for tickets := range ticketChan {
			log.Println(tickets)
		}
	}()

	select {}
}
