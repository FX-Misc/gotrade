package gotrade

import (
	"github.com/ideawebcn/gotrade"
	"testing"
)

var account *gotrade.Xueqiu

func init() {
	account = gotrade.NewXueqiu("/config/xueqiu.yaml")
}

func Test_Login(t *testing.T) {
	account.Login()
}

// func Test_Buy(t *testing.T) {
// 	account.Buy("150260", 1.430, 100)
// }
