package gotrade

import (
	"log"
	"testing"
    "sync"
)

func TestBackTestingSubscriber(t *testing.T) {
    wg := new(sync.WaitGroup)
	sbrBackTesting, err := NewBackTestingSubscriber("/var/stock_data", "2016-05-06", wg)
	if err != nil {
		panic(err)
	}
	codeList := []string{"150299"}
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
