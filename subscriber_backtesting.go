package gotrade

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SubscriberBackTesting struct {
	logger                   *logrus.Logger
	wg                       *sync.WaitGroup
	cacheQuotationLocker     *sync.RWMutex
	quotationCodeList        []string
	ticketCodeList           []string
	quotationCodeFound       map[string]bool
	ticketCodeFound          map[string]bool
	dateStr                  string
	dataFile                 string                     // path to data file
	quotationChanMap         map[string]chan *Quotation // quotation strategy to channel [strategyName]quotitaionChannel
	quotationCodeStrategyMap map[string][]string        // quotation code to strategies [code][]{strategy_1, strategy_2}
	cacheQuotaionChan        chan *Quotation            // for cache quotation
	quotationCacheMap        map[string]*Quotation      // cached quotation
	ticketChanMap            map[string]chan *Tickets   // ticket strategy to channel [strategyName]TicketChannel
	ticketCodeStrategyMap    map[string][]string        // ticket code to strategies [code][]{strategy_1, strategy_2}
}

func NewBackTestingSubscriber(storePath string, dateStr string, wg *sync.WaitGroup) (*SubscriberBackTesting, error) {
	sbr := new(SubscriberBackTesting)
	sbr.wg = wg
	sbr.quotationCodeList = []string{}
	sbr.ticketCodeList = []string{}
	sbr.quotationCodeStrategyMap = make(map[string][]string)
	sbr.quotationChanMap = make(map[string]chan *Quotation)
	sbr.ticketChanMap = make(map[string]chan *Tickets)
	sbr.ticketCodeStrategyMap = make(map[string][]string)
	sbr.cacheQuotaionChan = make(chan *Quotation)
	sbr.quotationCodeFound = make(map[string]bool)
	sbr.ticketCodeFound = make(map[string]bool)
	sbr.quotationCacheMap = make(map[string]*Quotation)
	sbr.logger = NewLogger("subscriber_backtesting")
	sbr.dateStr = dateStr
	sbr.dataFile = fmt.Sprintf("%s/%s.txt", storePath, dateStr)
	sbr.cacheQuotationLocker = new(sync.RWMutex)
	f, err := os.Open(sbr.dataFile)
	if err != nil {
		return nil, fmt.Errorf("data file %s not exists", sbr.dataFile)
	}
	f.Close()
	return sbr, nil
}

// 订阅基本行情
// id 通常为 strategyName-accountName
func (sbr *SubscriberBackTesting) Subscribe(id string, codes []string) (quotationChan chan *Quotation) {
	sbr.logger.Infof("subscribe strategy: %s code list: %q", id, codes)
	found := make(map[string]bool)
	for _, code := range codes {
		if !found[code] {
			found[code] = true
			sbr.quotationCodeStrategyMap[code] = append(sbr.quotationCodeStrategyMap[code], id)
		}
		if !sbr.quotationCodeFound[code] {
			sbr.quotationCodeFound[code] = true
			sbr.quotationCodeList = append(sbr.quotationCodeList, code)
		}
	}

	// 限制管道长度
	quotationChan = make(chan *Quotation, 100)
	sbr.quotationChanMap[id] = quotationChan
	return
}

// 订阅逐笔
func (sbr *SubscriberBackTesting) SubscribeTicket(id string, codes []string) (ticketChan chan *Tickets) {
	sbr.logger.Infof("subscribeTicket strategy: %s code list: %q", id, codes)
	found := make(map[string]bool)
	for _, code := range codes {
		if !found[code] {
			found[code] = true
			sbr.ticketCodeStrategyMap[code] = append(sbr.ticketCodeStrategyMap[code], id)
		}
		if !sbr.ticketCodeFound[code] {
			sbr.ticketCodeFound[code] = true
			sbr.ticketCodeList = append(sbr.ticketCodeList, code)
		}
	}

	// 限制管道长度
	ticketChan = make(chan *Tickets, 100)
	sbr.ticketChanMap[id] = ticketChan
	return
}

func (sbr *SubscriberBackTesting) Run() {
	go sbr.CacheQuotation()
	f, err := os.Open(sbr.dataFile)
	if err != nil {
		sbr.logger.Errorf("open data file %s failed %s", sbr.dataFile, err)
		return
	}
	scanner := bufio.NewScanner(f)
	sbr.logger.Printf("start scan data file %s", sbr.dataFile)
	for scanner.Scan() {
		raw := scanner.Text()
		if raw[0:2] == "0=" {
			sbr.decodeQuotation(raw[2:])
		} else if raw[0:2] == "1=" {
			sbr.decodeTickets(raw[2:])
		}
	}
	if err = scanner.Err(); err != nil {
		sbr.logger.Warningf("scan error %s", err)
	}
	f.Close()
	sbr.logger.Println("all data parsed")
	for _, ticketChan := range sbr.ticketChanMap {
		close(ticketChan)
	}

	for _, quotationChan := range sbr.quotationChanMap {
		close(quotationChan)
	}
	close(sbr.cacheQuotaionChan)
	sbr.wg.Done()
	return
}

func (sbr *SubscriberBackTesting) CacheQuotation() {
	for {
		select {
		case quotation, ok := <-sbr.cacheQuotaionChan:
			if !ok {
				return
			}
			sbr.cacheQuotationLocker.Lock()
			sbr.quotationCacheMap[quotation.Code] = quotation
			sbr.cacheQuotationLocker.Unlock()
		}
	}
}

func (sbr *SubscriberBackTesting) GetQuotation(code string) (quotation *Quotation, err error) {
	var found bool
	sbr.cacheQuotationLocker.Lock()
	quotation, found = sbr.quotationCacheMap[code]
	if !found {
		err = fmt.Errorf("%s not coming", code)
	}
	sbr.cacheQuotationLocker.Unlock()
	return
}

/*
0=code,time|close|pre_close|volume|trade_amount|bid1_price|bid2_price|....|bid1_amount|bid2_amount|...|ask1_price|ask2_price|...|ask1_amount|ask2_amount|...
*/
func (sbr *SubscriberBackTesting) decodeQuotation(raw string) {
	rawLines := strings.Split(raw, ",")
	if len(rawLines) != 2 {
		sbr.logger.Warningf("decode line %s failed", raw)
		return
	}
	quo := new(Quotation)
	quo.Code = rawLines[0]
	rawLines = strings.Split(rawLines[1], "|")
	if len(rawLines) != 45 {
		sbr.logger.Warningf("decode line %s failed", rawLines[1])
		return
	}
	quo.Time, _ = time.Parse("2006-01-02T15:04:05.999999-07:00", fmt.Sprintf("%sT%s+08:00", sbr.dateStr, rawLines[0]))
	quo.Now = quo.Time
	quo.Close, _ = strconv.ParseFloat(rawLines[1], 64)
	quo.PreClose, _ = strconv.ParseFloat(rawLines[2], 64)
	quo.Volume, _ = strconv.ParseFloat(rawLines[3], 64)
	quo.TradeAmount, _ = strconv.ParseFloat(rawLines[4], 64)
	rawLines = rawLines[5:]
	quo.Bids = make([]OrderBook, 10)
	quo.Asks = make([]OrderBook, 10)
	for index := 0; index < 10; index++ {
		quo.Bids[index].Price, _ = strconv.ParseFloat(rawLines[index], 64)
		quo.Bids[index].Amount, _ = strconv.ParseFloat(rawLines[10+index], 64)
		quo.Asks[index].Price, _ = strconv.ParseFloat(rawLines[20+index], 64)
		quo.Asks[index].Amount, _ = strconv.ParseFloat(rawLines[30+index], 64)
	}

	// push
	sbr.cacheQuotaionChan <- quo
	strategyNameList := sbr.quotationCodeStrategyMap[quo.Code]
	for _, strategyName := range strategyNameList {
		sbr.quotationChanMap[strategyName] <- quo
	}
}

/*
1=code,time|price|amount|type,time|price|amount|type
*/
func (sbr *SubscriberBackTesting) decodeTickets(raw string) {
	rawLines := strings.Split(raw, ",")
	if len(raw) < 2 {
		sbr.logger.Printf("empty tickets line %s", raw)
		return
	}
	tickets := new(Tickets)
	tickets.Code = rawLines[0]
	rawLines = rawLines[1:]
	for _, ticketLine := range rawLines {
		ticketLines := strings.Split(ticketLine, "|")
		if len(ticketLines) != 4 {
			sbr.logger.Warningf("parse ticket line %s failed", ticketLine)
			continue
		}
		ticket := Ticket{}
		ticket.Time, _ = time.Parse("2006-01-02T15:04:05.999999-07:00", fmt.Sprintf("%sT%s+08:00", sbr.dateStr, ticketLines[0]))
		ticket.Price, _ = strconv.ParseFloat(ticketLines[1], 64)
		ticket.Amount, _ = strconv.ParseFloat(ticketLines[2], 64)
		if ticketLines[3] == "0" {
			ticket.Type = TICKET_SELL
		} else if ticketLines[3] == "2" {
			ticket.Type = TICKET_BUY
		}
		tickets.Tickets = append(tickets.Tickets, ticket)
	}
	tickets.Now = tickets.Tickets[0].Time
	// push
	strategyNameList := sbr.ticketCodeStrategyMap[tickets.Code]
	for _, strategyName := range strategyNameList {
		sbr.ticketChanMap[strategyName] <- tickets
	}
}
