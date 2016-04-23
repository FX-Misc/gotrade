package gotrade

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
)

type AccountBacktesting struct {
	logger           *logrus.Logger
	FeeValue         float64 `yaml:"fee"`
	Nickname         string  `yaml:nickname`
	Username         string  `yaml:"username"`
	TotalBalance     float64 `yaml:"balance"`
	MarketBalance    float64 `yaml:"market_balance"`
	AvailableBalance float64 `yaml:"available_balance"`
	FrozenBalance    float64 `yaml:"frozen_balance"`
}

func NewTestAccount(configPath string) (account *AccountBacktesting) {
	account = &AccountBacktesting{}
	account.logger = NewLogger("trader")
	err := YamlFileDecode(configPath, account)
	if err != nil {
		account.logger.Error("init account error")
		panic(err)
	}
	return
}

func (account *AccountBacktesting) Login() (err error) {
	account.logger.Printf("%s login success", account.Nickname)
	return
}

func (account *AccountBacktesting) Name() string {
	return account.Nickname
}

func (account *AccountBacktesting) Fee() float64 {
	return account.FeeValue
}

func (account *AccountBacktesting) Buy(code string, price float64, amount float64) (id int64, err error) {
	requireMoney := price * amount * (1 + account.FeeValue)
	if requireMoney > account.AvailableBalance {
		return 0, fmt.Errorf("buy error,no money")
	}
	account.AvailableBalance -= requireMoney
	account.TotalBalance -= requireMoney
	account.updatePosition(code, price, amount)
	YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/account.yaml", account)
	return time.Now().Unix(), nil
}

func (account *AccountBacktesting) Sell(code string, price float64, amount float64) (id int64, err error) {
	requireMoney := price * amount * (1 - account.FeeValue)
	position := account.getPosition(code)
	if position == nil || position.AvailableAmount < amount {
		return 0, fmt.Errorf("sell error,no stock")
	}
	account.AvailableBalance += requireMoney
	account.TotalBalance += requireMoney
	account.updatePosition(code, price, -amount)
	YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/account.yaml", account)
	return time.Now().Unix(), nil
}

func (account *AccountBacktesting) Cancel(id int64) (err error) {
	return fmt.Errorf("already deal")
}

func (account *AccountBacktesting) Position() (data []*StockPosition, err error) {
	positions := []BackTestingPosition{}
	YamlFileDecode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", &positions)
	for _, position := range positions {
		stockPosition := new(StockPosition)
		stockPosition.Name = position.Code
		stockPosition.Code = position.Code
		stockPosition.Amount = position.Amount
		stockPosition.AvailableAmount = position.AvailableAmount
		// 以下是@todo
		stockPosition.FrozenAmount = 0
		stockPosition.Total = 1
		stockPosition.Profit = 1
		stockPosition.ProfitRatio = 1
		stockPosition.ProfitRatio /= 100
		data = append(data, stockPosition)
	}
	return
}

func (account *AccountBacktesting) GetPositionMap() (positionMap map[string]*StockPosition, err error) {
	positionList, err := account.Position()
	if err != nil {
		return
	}
	positionMap = make(map[string]*StockPosition)
	for _, position := range positionList {
		positionMap[position.Code] = position
	}
	return
}

func (account *AccountBacktesting) GetPendingMap() (orderMap map[string]*Order, err error) {
	orderList, err := account.Pending()
	if err != nil {
		return
	}
	orderMap = make(map[string]*Order)
	for _, order := range orderList {
		orderMap[order.Code] = &order
	}
	return
}

func (account *AccountBacktesting) Balance() (data Balance, err error) {
	data.Balance = account.TotalBalance
	data.MarketBalance = account.MarketBalance
	data.AvailableBalance = account.AvailableBalance
	data.FrozenBalance = account.FrozenBalance
	return
}

func (account *AccountBacktesting) Pending() (data []Order, err error) {
	return
}

func (account *AccountBacktesting) DeferCancel(id int64, afterSecond int64) {
	return
}

// 回测时，每日过后需要运行这个方法
func (account *AccountBacktesting) DailyUpdate() {
	positions := []BackTestingPosition{}
	YamlFileDecode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", &positions)
	for _, position := range positions {
		position.AvailableAmount = position.Amount
		YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", positions)
	}
	YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", positions)
}

type BackTestingPosition struct {
	Code            string  `yaml:"code"`
	Price           float64 `yaml:"price"`
	Amount          float64 `yaml:"amount"`
	AvailableAmount float64 `yaml:"available_amount"`
}

func (account *AccountBacktesting) updatePosition(code string, price float64, amount float64) {
	positions := []BackTestingPosition{}
	YamlFileDecode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", &positions)

	for _, position := range positions {
		if position.Code == code {
			position.Price = (position.Price*position.Amount + price*amount) / (position.Amount + amount)
			position.Amount += amount
			if amount < 0 {
				position.AvailableAmount += amount
			}
			YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", positions)
			return
		}
	}

	positions = append(positions, BackTestingPosition{
		Code:   code,
		Price:  price,
		Amount: amount,
	})

	YamlFileEncode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", positions)
}

func (account *AccountBacktesting) getPosition(code string) *BackTestingPosition {
	positions := []BackTestingPosition{}
	YamlFileDecode(GetBasePath()+"/config/"+account.Username+"/positions.yaml", &positions)

	for _, position := range positions {
		if position.Code == code {
			return &position
		}
	}
	return nil
}
