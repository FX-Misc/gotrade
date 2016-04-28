package gotrade

type StockPosition struct {
	Code            string  `yaml:"code"`
	Name            string  `yaml:"name"`
	Amount          float64 `yaml:"amount"` //数量
	Total           float64 `yaml:"total"`  // 价值
	AvailableAmount float64 `yaml:"available_amount"`
	FrozenAmount    float64 `yaml:"frozen_amount"`
	Profit          float64 `yaml:"profit"`       // 盈利
	ProfitRatio     float64 `yaml:"profit_ratio"` // 盈利率
}

type Balance struct {
	Balance          float64
	MarketBalance    float64
	AvailableBalance float64
	FrozenBalance    float64
}

type Order struct {
	Code   string
	Name   string
	Amount float64
	Price  float64
	Id     int64
	Type   string
}

type Account interface {
	Login() error
	Name() string
	Fee() float64
	Buy(string, float64, float64) (int64, error)
	Sell(string, float64, float64) (int64, error)
	Cancel(int64) error
	DeferCancel(int64, int64)
	Position() ([]*StockPosition, error)
	Balance() (Balance, error)
	Pending() ([]Order, error)
	GetPositionMap() (positionMap map[string]*StockPosition, err error)
}
