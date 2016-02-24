package gotrade

import (
	"gotrade"
	"testing"
)

var account *gotrade.Account

func init() {
	account = gotrade.NewAccount("config/trade.yaml")
}

func Test_Login(t *testing.T) {
	account.Login()
}

func Test_Postion(t *testing.T) {
	t.Log(account.Position())
	t.Log(account.Balance())

}

// func Test_Buy(t *testing.T) {
// 	a := newAccount()
// 	id, err := a.Buy("150260", 1.430, 100)
// 	log.Println(id, err)
// }
