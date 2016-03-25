package gotrade

import (
	"github.com/ideawebcn/gotrade"
	"log"
	"testing"
)

var sbr *gotrade.Subscriber

func init() {
	sbr = gotrade.NewSubscriber("config/subscriber.yaml")
}

func Test_Run(t *testing.T) {
	codeList := []string{"150168", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250", "150167", "150250"}
	quoChan := sbr.Subscribe("test", codeList)
	sbr.Unsubscribe("test", []string{"150168"})

	go sbr.Run()
	go func() {
		for quo := range quoChan {
			log.Println(quo)
		}
	}()

	select {}
}
