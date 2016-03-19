package gotrade

import (
	"github.com/ideawebcn/gotrade"
	"log"
	"testing"
)

var sbr *gotrade.Subscriber

func init() {
	sbr = gotrade.NewSubscriber("config/subscribe.yaml")
}

func Test_Run(t *testing.T) {
	codeList := []string{"150168", "150167", "150250"}
	quoChan := sbr.Subscribe("test", codeList)
	quoChan2 := sbr.Subscribe("tttt", codeList)
	sbr.Unsubscribe("test", []string{"150168"})

	go sbr.Run()
	go func() {
		for quo := range quoChan {
			log.Println(quo)
		}
	}()
	go func() {
		for quo := range quoChan2 {
			log.Println(quo)
		}
	}()

	select {}
}
