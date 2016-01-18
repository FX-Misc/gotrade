package gotrade

import (
	"gotrade"
	"log"
	"testing"
)

func newSubscriber() *gotrade.Subscriber {
	subscriber := gotrade.NewSubscriber("./config/subscribe.yaml")
	return subscriber
}

func Test_Run(t *testing.T) {
	subr := newSubscriber()
	codeList := []string{"600000", "502028", "150167"}
	quoChan := subr.Subscribe("test", codeList)
	go subr.Run()
	for quo := range quoChan {
		log.Println(quo)
	}
}
