package gotrade

import (
	"gotrade"
	"log"
	"testing"
)

func newAccount() *gotrade.Account {
	account := &gotrade.Account{}
	err := gotrade.YamlFileDecode("./config/trade.yaml", account)
	if err != nil {
		panic(err)
	}
	account.Login()
	return account
}

func Test_Login(t *testing.T) {
	a := newAccount()
	a.Login()
}

func Test_Postion(t *testing.T) {
	a := newAccount()
	log.Println(a.Position())
}

func Test_Buy(t *testing.T) {
	a := newAccount()
	id, err := a.Buy("150260", 1.430, 100)
	log.Println(id, err)
}
