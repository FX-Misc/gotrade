package gotrade

import (
	"log"
	"testing"
)

func TestBackTestingSubscriber(t *testing.T) {
	sbrBackTesting := NewBackTestingSubscriber("/tmp/storage", "2016-04-22", "2016-04-22")
	codeList := []string{"002775"}
	quoChan := sbrBackTesting.Subscribe("test", codeList)
	ticketChan := sbrBackTesting.SubscribeTicket("test", codeList)

	go sbrBackTesting.Run()
	go func() {
		for quotation := range quoChan {
			log.Println(quotation)
		}
	}()

	go func() {
		for tickets := range ticketChan {
			log.Println(tickets)
		}
	}()

	select {}
}
