package gotrade

import (
	"fmt"
	"time"
)

const (
	TICKET_BUY  = 0
	TICKET_SELL = 1
)

type Subscriber interface {
	Subscribe(string, []string) chan *Quotation
	SubscribeTicket(string, []string) chan *Tickets
	GetQuotation(code string) (*Quotation, error)
}

type QuotationStack struct {
	Length     int
	quotations []*Quotation
}

type OrderBook struct {
	Price  float64
	Amount float64
}

type Tickets struct {
	Code    string
	Now     time.Time
	Tickets []Ticket
}

type Ticket struct {
	Price  float64
	Amount float64
	Type   int
	Time   time.Time
}

type Configuration struct {
	Cookie      string `yaml:"cookie"`
	UA          string `yaml:"ua"`
	TokenServer string `yaml:"token_server"`
}

type Quotation struct {
	Code        string
	Name        string
	PreClose    float64
	Close       float64
	Volume      float64 // 成交量(元)
	TradeAmount float64 // 成交股数
	Time        time.Time
	Now         time.Time
	Bids        []OrderBook
	Asks        []OrderBook
}

func (q *Quotation) GetDepthPrice(minDepth float64, side string) float64 {
	// 获取深度为depth的出价 也就是卖价
	depth := 0.0
	if side == "bid" || side == "sell" {
		for _, bid := range q.Bids {
			depth += bid.Amount
			if depth >= minDepth {
				return bid.Price
			}
		}
		return q.Bids[0].Price
	} else {
		// 买价
		for _, ask := range q.Asks {
			depth += ask.Amount
			if depth >= minDepth {
				return ask.Price
			}
		}
		return q.Asks[0].Price
	}
}

func (qs *QuotationStack) Push(q *Quotation) error {
	if qs.Length <= 0 {
		return fmt.Errorf("unexpected length %d", qs.Length)
	}
	qs.quotations = append(qs.quotations, q)
	curLen := len(qs.quotations)
	if curLen == qs.Length {
		return nil
	}
	if curLen > qs.Length {
		qs.quotations = qs.quotations[curLen-qs.Length : curLen]
	}
	return nil
}

func (qs *QuotationStack) All() ([]*Quotation, error) {
	if qs.Length != len(qs.quotations) {
		return []*Quotation{}, fmt.Errorf("stack required %d but is %d", qs.Length, len(qs.quotations))
	}
	return qs.quotations, nil
}
