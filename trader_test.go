package gotrade

import (
	"testing"
)

var account *Account

func init() {
	account = NewAccount("config/trade.yaml")
}

func Test_Login(t *testing.T) {
	account.Login()
}

func Test_Postion(t *testing.T) {
	t.Log(account.Position())
}

// func Test_Buy(t *testing.T) {
// 	a := newAccount()
// 	id, err := a.Buy("150260", 1.430, 100)
// 	log.Println(id, err)
// }
