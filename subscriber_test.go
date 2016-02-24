package gotrade

import (
	"gotrade"
	"log"
	"testing"
)

var sbr *gotrade.Subscriber

func init() {
	sbr = gotrade.NewSubscriber("config/subscribe.yaml")
}

func Test_Run(t *testing.T) {
	codeList := []string{"150168", "150167"}
	quoChan := sbr.Subscribe("test", codeList)
	go sbr.Run()
	for quo := range quoChan {
		log.Println(quo)
	}
}
