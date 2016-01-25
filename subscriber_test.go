package gotrade

import (
	"testing"
)

var sbr *Subscriber

func init() {
	sbr = NewSubscriber("config/subscribe.yaml")
}

func TestSubscribe(t *testing.T) {
	codeList := []string{"600000", "502028", "150167"}
	quoChan := sbr.Subscribe("test", codeList)
	go sbr.Run()
	for quo := range quoChan {
		t.Log(quo)
	}
}
