package gotrade

import (
	"testing"
)

var account *AccountHuatai

func init() {
	account = NewHuataiAccount("config/trade.yaml")
}

func Test_Login(t *testing.T) {
	account.Login()
}

func Test_Postion(t *testing.T) {
	t.Log(account.Pending())
	t.Log(account.Balance())

}

// func Test_Buy(t *testing.T) {
// 	account.Buy("150260", 1.430, 100)
// }
