package gotrade

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SubscriberSina struct {
	TokenServer              string
	IP                       string
	Cookie                   string
	UA                       string
	today                    string
	logger                   *logrus.Logger
	quotationCodeList        []string
	ticketCodeList           []string
	quotationCodeFound       map[string]bool
	ticketCodeFound          map[string]bool
	quotationChanMap         map[string]chan *Quotation // quotation strategy to channel [strategyName]quotitaionChannel
	quotationCodeStrategyMap map[string][]string        // quotation code to strategies [code][]{strategy_1, strategy_2}
	cacheQuotaionChan        chan *Quotation            // for cache quotation
	quotationCacheMap        map[string]*Quotation      // cached quotation
	ticketChanMap            map[string]chan *Tickets   // ticket strategy to channel [strategyName]TicketChannel
	ticketCodeStrategyMap    map[string][]string        // ticket code to strategies [code][]{strategy_1, strategy_2}
}

type Api struct {
	Params             string
	flag               int
	tokenExpired       bool
	token              string
	subscriber         *SubscriberSina
	quotationTimeCache map[string]int64
}

func NewSinaSubscriber(configPath string) (subscriber *SubscriberSina) {
	config := &Configuration{}
	err := YamlFileDecode(configPath, config)
	if err != nil {
		panic(err)
	}
	subscriber = &SubscriberSina{}
	subscriber.quotationCodeList = []string{}
	subscriber.ticketCodeList = []string{}
	subscriber.quotationCodeStrategyMap = make(map[string][]string)
	subscriber.quotationChanMap = make(map[string]chan *Quotation)
	subscriber.ticketChanMap = make(map[string]chan *Tickets)
	subscriber.ticketCodeStrategyMap = make(map[string][]string)
	subscriber.cacheQuotaionChan = make(chan *Quotation)
	subscriber.Cookie = config.Cookie
	subscriber.UA = config.UA
	subscriber.quotationCodeFound = make(map[string]bool)
	subscriber.ticketCodeFound = make(map[string]bool)
	err = subscriber.getExternalIp()
	if err != nil {
		panic(err)
	}
	subscriber.quotationCacheMap = make(map[string]*Quotation)
	subscriber.logger = NewLogger("subscriber")
	subscriber.logger.Infof("external IP is %s", subscriber.IP)
	subscriber.TokenServer = config.TokenServer
	subscriber.today = time.Now().Format("2006-01-02")
	return
}

// 订阅基本行情
func (s *SubscriberSina) Subscribe(strategyName string, codeList []string) (quotationChan chan *Quotation) {
	s.logger.Infof("subscribe strategy: %s code list: %q", strategyName, codeList)
	found := make(map[string]bool)
	for _, code := range codeList {
		if !found[code] {
			found[code] = true
			s.quotationCodeStrategyMap[code] = append(s.quotationCodeStrategyMap[code], strategyName)
		}
		if !s.quotationCodeFound[code] {
			s.quotationCodeFound[code] = true
			s.quotationCodeList = append(s.quotationCodeList, code)
		}
	}
	quotationChan = make(chan *Quotation)
	s.quotationChanMap[strategyName] = quotationChan
	return
}

// 订阅逐笔
func (s *SubscriberSina) SubscribeTicket(strategyName string, codeList []string) (ticketChan chan *Tickets) {
	s.logger.Infof("subscribeTicket strategy: %s code list: %q", strategyName, codeList)
	found := make(map[string]bool)
	for _, code := range codeList {
		if !found[code] {
			found[code] = true
			s.ticketCodeStrategyMap[code] = append(s.ticketCodeStrategyMap[code], strategyName)
		}
		if !s.ticketCodeFound[code] {
			s.ticketCodeFound[code] = true
			s.ticketCodeList = append(s.ticketCodeList, code)
		}
	}
	ticketChan = make(chan *Tickets)
	s.ticketChanMap[strategyName] = ticketChan
	return
}

func (sbr *SubscriberSina) Run() {
	sbr.logger.Info("running...")
	log.Printf("subscribe quotation list %s", sbr.quotationCodeList)
	log.Printf("subscribe ticket list %s", sbr.ticketCodeList)

	// cache quation
	go func() {
		for quotation := range sbr.cacheQuotaionChan {
			sbr.quotationCacheMap[quotation.Code] = quotation
		}
	}()

	quotationParamList := make([]string, 0)
	for _, code := range sbr.quotationCodeList {
		if code[0:2] == "15" || code[0:2] == "00" || code[0:2] == "30" {
			quotationParamList = append(quotationParamList, fmt.Sprintf("2cn_sz%s", code))
			continue
		}
		if code[0:2] == "60" || code[0:1] == "5" {
			quotationParamList = append(quotationParamList, fmt.Sprintf("2cn_sh%s", code))
			continue
		}
		// index code
		if code[0:3] == "i39" {
			quotationParamList = append(quotationParamList, fmt.Sprintf("sz%s", code[1:7]))
			continue
		}
		if code[0:3] == "i00" {
			quotationParamList = append(quotationParamList, fmt.Sprintf("sh%s", code[1:7]))
			continue
		}
	}

	ticketParamList := make([]string, 0)
	for _, code := range sbr.ticketCodeList {
		if code[0:2] == "15" || code[0:2] == "00" || code[0:2] == "30" {
			ticketParamList = append(ticketParamList, fmt.Sprintf("2cn_sz%s_0,2cn_sz%s_1", code, code))
			continue
		}
		if code[0:2] == "60" || code[0:1] == "5" {
			ticketParamList = append(ticketParamList, fmt.Sprintf("2cn_sh%s_0,2cn_sh%s_1", code, code))
			continue
		}
	}

	// subscribe quotation
	start, end, flag := 0, 0, 0
	length := len(quotationParamList)
	if length > 0 {
		for {
			end = start + 62
			if end >= length {
				end = length
			}
			params := strings.Join(quotationParamList[start:end], ",")
			api := Api{
				Params:             params,
				subscriber:         sbr,
				flag:               flag,
				quotationTimeCache: make(map[string]int64, 0),
			}
			flag += 1
			go api.Run()
			// 分散 worker 防止并发量过大
			time.Sleep(time.Second * 1)
			if end >= length {
				break
			}
			start = start + 62
		}
	}

	// subscribe ticket
	start, end = 0, 0
	length = len(ticketParamList)
	if length > 0 {
		for {
			end = start + 30
			if end >= length {
				end = length
			}
			params := strings.Join(ticketParamList[start:end], ",")
			api := Api{
				Params:     params,
				subscriber: sbr,
				flag:       flag,
			}
			flag += 1
			go api.Run()
			// 分散 worker 防止并发量过大
			time.Sleep(time.Second * 1)
			if end >= length {
				break
			}
			start = start + 30
		}
	}

	log.Printf("all %d worker started", flag)
}

func (api *Api) Run() {
	err := api.refreshToken()
	if err != nil {
		log.Printf("#%d %s", api.flag, err)
		api.tokenExpired = true
	} else {
		api.tokenExpired = false
	}
	go func() {
		lastRefresh := time.Now().Unix()
		for {
			time.Sleep(time.Millisecond * 500)

			// 如果 token 过期 或者离上次刷新超过5分钟，刷新 token
			if api.tokenExpired || time.Now().Unix()-lastRefresh > 300 {
				lastRefresh = time.Now().Unix()
				err := api.refreshToken()
				if err != nil {
					log.Printf("#%d %s", api.flag, err)
					api.tokenExpired = true
					continue
				}
				api.tokenExpired = false
			}
		}
	}()

	for {
		// token 已过期，等待重连
		if api.tokenExpired {
			time.Sleep(time.Second * 1)
			log.Printf("#%d waiting token...", api.flag)
			continue
		}
		log.Printf("#%d connect", api.flag)
		err := api.connect()
		if err != nil {
			log.Printf("#%d connect failed: %s", api.flag, err)
		}
		log.Printf("#%d closed", api.flag)
	}
}

// deprecated
func (s *SubscriberSina) Unsubscribe(strategyName string, codeList []string) (err error) {
	s.logger.Infof("unsubscribe strategy: %s code list: %q", strategyName, codeList)
	for _, code := range codeList {
		for i, name := range s.quotationCodeStrategyMap[code] {
			if name == strategyName {
				s.quotationCodeStrategyMap[code] = append(s.quotationCodeStrategyMap[code][:i], s.quotationCodeStrategyMap[code][i+1:]...)
			}
		}
	}
	return
}

func (s *SubscriberSina) getExternalIp() error {
	resp, err := http.Get("https://ff.sinajs.cn/?_=1453697286404&list=sys_clientip")
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return err
	}
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	result := re.FindAllString(string(string(b)), -1)
	if len(result) == 1 {
		s.IP = result[0]
		return nil
	}
	return fmt.Errorf("get external ip failed")
}

func (api *Api) connect() error {
	url := fmt.Sprintf("ws://ff.sinajs.cn/wskt?token=%s&list=%s", api.token, api.Params)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// keep alive
	keepAliveTicker := time.NewTicker(1 * time.Minute)
	sendTokenTicker := time.NewTicker(3 * time.Minute)
	destoryChan := make(chan bool)
	defer keepAliveTicker.Stop()
	defer sendTokenTicker.Stop()
	defer close(destoryChan)
	go func() {
		for {
			select {
			case <-keepAliveTicker.C:
				err := c.WriteMessage(1, []byte(""))
				if err != nil {
					log.Printf("#%d send empty message failed: %s", api.flag, err)
				}
			case <-sendTokenTicker.C:
				log.Printf("#%d send token %s", api.flag, api.token)
				err := c.WriteMessage(1, []byte("*"+api.token))
				if err != nil {
					log.Printf("#%d send token failed: %s", api.flag, err)
				}
			case <-destoryChan:
				log.Printf("#%d destroy connection worker", api.flag)
				return
			}
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			destoryChan <- true
			return err
		}
		raw := string(message)
		if strings.Contains(raw, "sys_auth=FAILED") {
			// 标记 token 为过期
			destoryChan <- true
			api.tokenExpired = true
			return fmt.Errorf("auth timeout")
		}
		rawLines := strings.SplitN(raw, "\n", -1)

		// @todo 如果有股票加入可能index的code会冲突
		for _, rawLine := range rawLines {
			if strings.Contains(rawLine, "sys_nxkey=") || strings.Contains(rawLine, "sys_time=") || strings.Contains(rawLine, "sys_auth=") || len(rawLine) < 10 {
				continue
			}
			// ticket
			if len(rawLine) > 15 && (rawLine[12:14] == "_0" || rawLine[12:14] == "_1") {
				tickets, err := api.parseTicket(rawLine)
				if err == nil {
					strategyNameList := api.subscriber.ticketCodeStrategyMap[tickets.Code]
					for _, strategyName := range strategyNameList {
						api.subscriber.ticketChanMap[strategyName] <- tickets
					}
				}
				continue
			}
			// quotation
			quo, err := api.parseQuotation(rawLine)
			if err == nil {
				if quo.TradeAmount == 0 && quo.Volume == 0 && quo.Close != quo.PreClose && quo.Close != 0 && quo.Bids[0].Price == quo.Asks[0].Price && quo.Bids[0].Amount == quo.Asks[0].Amount {
					quo.Code = "i" + quo.Code
				}

				// filter duplicate quotation by remote time
				if timestamp, found := api.quotationTimeCache[quo.Code]; found && timestamp == quo.Time.Unix() {
					continue
				} else {
					api.quotationTimeCache[quo.Code] = quo.Time.Unix()
				}

				api.subscriber.cacheQuotaionChan <- quo
				strategyNameList := api.subscriber.quotationCodeStrategyMap[quo.Code]
				for _, strategyName := range strategyNameList {
					api.subscriber.quotationChanMap[strategyName] <- quo
				}
			}
		}
	}
}

func (api *Api) parseQuotation(rawLine string) (*Quotation, error) {
	rawLines := strings.SplitN(rawLine, "=", 2)
	if len(rawLines) < 2 {
		return nil, fmt.Errorf("unexpected data %s", rawLine)
	}
	quo := new(Quotation)
	if rawLines[0][0:4] == "2cn_" {
		quo.Code = rawLines[0][6:12]
		rawLines = strings.Split(rawLines[1], ",")
		length := len(rawLines)
		if length < 40 {
			return nil, fmt.Errorf("unexpected data %s", rawLine)
		}
		quo.Name = rawLines[0]
		quo.Time, _ = time.Parse("2006-01-02T15:04:05.999999-07:00", fmt.Sprintf("%sT%s+08:00", rawLines[2], rawLines[1]))
		quo.Now = time.Now()
		quo.PreClose, _ = strconv.ParseFloat(rawLines[3], 64)
		quo.Close, _ = strconv.ParseFloat(rawLines[7], 64)
		quo.TradeAmount, _ = strconv.ParseFloat(rawLines[10], 64)
		quo.Volume, _ = strconv.ParseFloat(rawLines[11], 64)
		rawLines = rawLines[length-40 : length]
		quo.Bids = make([]OrderBook, 10)
		quo.Asks = make([]OrderBook, 10)
		for index := 0; index < 10; index++ {
			quo.Bids[index].Price, _ = strconv.ParseFloat(rawLines[index], 64)
			quo.Bids[index].Amount, _ = strconv.ParseFloat(rawLines[10+index], 64)
			quo.Asks[index].Price, _ = strconv.ParseFloat(rawLines[20+index], 64)
			quo.Asks[index].Amount, _ = strconv.ParseFloat(rawLines[30+index], 64)
		}
		return quo, nil
	} else {
		if len(rawLines) < 2 || len(rawLines[1]) < 10 {
			return nil, fmt.Errorf("unable parse data")
		}
		quo.Code = rawLines[0][2:8]
		rawLines = strings.Split(rawLines[1], ",")
		if len(rawLines) < 32 {
			return nil, fmt.Errorf("unable parse data")
		}
		quo.Name = rawLines[0]
		quo.Time, _ = time.Parse("2006-01-02T15:04:05.999999-07:00", fmt.Sprintf("%sT%s+08:00", rawLines[30], rawLines[31]))
		quo.Now = time.Now()
		quo.PreClose, _ = strconv.ParseFloat(rawLines[2], 64)
		quo.Close, _ = strconv.ParseFloat(rawLines[3], 64)
		quo.Bids = make([]OrderBook, 10)
		quo.Asks = make([]OrderBook, 10)
		return quo, nil
	}
}

func (api *Api) parseTicket(rawLine string) (*Tickets, error) {
	rawLines := strings.SplitN(rawLine, "=", 2)
	if len(rawLines) < 2 {
		return nil, fmt.Errorf("unexpected data %s", rawLine)
	}
	tickets := new(Tickets)
	tickets.Code = rawLines[0][6:12]
	rawLines = strings.Split(rawLines[1], ",")
	for _, rawTicket := range rawLines {
		ticketSlice := strings.Split(rawTicket, "|")
		if len(ticketSlice) != 9 {
			return nil, fmt.Errorf("unexpected data %s", rawTicket)
		}
		ticket := Ticket{}
		ticket.Amount, _ = strconv.ParseFloat(ticketSlice[3], 64)
		ticket.Price, _ = strconv.ParseFloat(ticketSlice[2], 64)
		ticket.Time, _ = time.Parse("2006-01-02T15:04:05.999999-07:00", fmt.Sprintf("%sT%s+08:00", api.subscriber.today, ticketSlice[1]))
		if ticketSlice[7] == "0" {
			ticket.Type = TICKET_SELL
		} else if ticketSlice[7] == "2" {
			ticket.Type = TICKET_BUY
		}
		tickets.Tickets = append(tickets.Tickets, ticket)
	}
	tickets.Now = time.Now()
	return tickets, nil
}

func (api *Api) refreshToken() error {
	log.Printf("#%d refresh token", api.flag)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	if !strings.Contains(api.Params, "2cn_") {
		api.Params = "2cn_sh510580," + api.Params
	}
	url := fmt.Sprintf(api.subscriber.TokenServer, api.subscriber.IP, api.Params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", api.subscriber.UA)
	req.Header.Add("Cookie", api.subscriber.Cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re := regexp.MustCompile(`result:"(.+?)"`)
	result := re.FindAllSubmatch(body, 1)
	if len(result) == 1 && len(result[0]) == 2 {
		token := string(result[0][1])
		if !strings.Contains(token, "have not buy") {
			api.token = token
			log.Printf("#%d get token %s", api.flag, api.token)
			return nil
		}
	}
	return fmt.Errorf("can't match token")
}

func (s *SubscriberSina) GetQuation(code string) (quotation *Quotation, err error) {
	var found bool
	quotation, found = s.quotationCacheMap[code]
	if !found {
		err = fmt.Errorf("%s not coming", code)
		quotation = &Quotation{}
	}
	return
}
