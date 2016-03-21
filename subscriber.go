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
	Volume      float64
	TradeAmount float64
	Time        time.Time
	Now         time.Time
	Bids        []OrderBook
	Asks        []OrderBook
}

type QuotationStack struct {
	Length     int
	quotations []*Quotation
}

type OrderBook struct {
	Price  float64
	Amount float64
}

type Subscriber struct {
	Params            string
	TokenServer       string
	IP                string
	Cookie            string
	UA                string
	token             string
	logger            *logrus.Logger
	codeList          []string
	quotationChanMap  map[string]chan *Quotation
	strategyMap       map[string][]string
	quotationCacheMap map[string]*Quotation
}

type Api struct {
	Params           string
	Cookie           string
	IP               string
	UA               string
	TokenServer      string
	token            string
	logger           *logrus.Logger
	subscriber       *Subscriber
	cache            map[string]*Quotation
	quotationChanMap map[string]chan *Quotation
	strategyMap      map[string][]string
}

func NewSubscriber(configPath string) (subscriber *Subscriber) {
	config := &Configuration{}
	err := YamlFileDecode(configPath, config)
	if err != nil {
		panic(err)
	}
	subscriber = &Subscriber{}
	subscriber.codeList = []string{}
	subscriber.strategyMap = make(map[string][]string)
	subscriber.quotationChanMap = make(map[string]chan *Quotation)
	subscriber.Cookie = config.Cookie
	subscriber.UA = config.UA
	err = subscriber.getExternalIp()
	if err != nil {
		panic(err)
	}
	subscriber.quotationCacheMap = make(map[string]*Quotation, 0)
	subscriber.logger = NewLogger("subscriber")
	subscriber.logger.Infof("external IP is %s", subscriber.IP)
	subscriber.TokenServer = config.TokenServer
	return
}

func (sbr *Subscriber) Run() {
	sbr.logger.Info("running...")
	found := make(map[string]bool)
	uniqueCodeList := []string{}
	for len(sbr.codeList) == 0 {
		sbr.logger.Info("no subscribe stock subscriber waiting...")
		time.Sleep(1 * time.Second)
	}
	for _, code := range sbr.codeList {
		if !found[code] {
			found[code] = true
			uniqueCodeList = append(uniqueCodeList, code)
		}
	}
	sbr.codeList = uniqueCodeList
	log.Printf("subscribe list %s", sbr.codeList)
	for key, _ := range sbr.codeList {
		if sbr.codeList[key][0:2] == "15" || sbr.codeList[key][0:2] == "00" || sbr.codeList[key][0:2] == "30" {
			sbr.codeList[key] = "2cn_sz" + sbr.codeList[key]
		}
		if sbr.codeList[key][0:2] == "60" || sbr.codeList[key][0:1] == "5" {
			sbr.codeList[key] = "2cn_sh" + sbr.codeList[key]
		}
		if sbr.codeList[key][0:3] == "i39" {
			sbr.codeList[key] = "sz" + sbr.codeList[key][1:7]
		}
		if sbr.codeList[key][0:3] == "i00" {
			sbr.codeList[key] = "sh" + sbr.codeList[key][1:7]
		}
	}

	start := 0
	end := 0
	length := len(sbr.codeList)

	for {
		end = start + 50
		if end >= length {
			end = length
		}
		params := strings.Join(sbr.codeList[start:end], ",")
		api := Api{
			Params:           params,
			IP:               sbr.IP,
			Cookie:           sbr.Cookie,
			UA:               sbr.UA,
			TokenServer:      sbr.TokenServer,
			quotationChanMap: sbr.quotationChanMap,
			strategyMap:      sbr.strategyMap,
			logger:           sbr.logger,
			subscriber:       sbr,
		}
		go api.Run()
		time.Sleep(time.Millisecond * 100)
		if end >= length {
			break
		}
		start = start + 50
	}
}

func (api *Api) Run() {
	api.refreshToken()
	go func() {
		for {
			time.Sleep(time.Minute * 1)
			api.refreshToken()
		}
	}()
	for {
		err := api.connect()
		if err != nil {
			log.Printf("connect failed: %s\n", err)
		}
		log.Println("closed")
	}
}

// @todo Run后实时新增订阅
func (s *Subscriber) Subscribe(strategyName string, codeList []string) (quotationChan chan *Quotation) {
	s.logger.Infof("subscribe strategy : %s code list : %q", strategyName, codeList)
	for _, code := range codeList {
		s.strategyMap[code] = append(s.strategyMap[code], strategyName)
	}
	s.codeList = append(s.codeList, codeList...)
	quotationChan = make(chan *Quotation)
	s.quotationChanMap[strategyName] = quotationChan
	return
}

func (s *Subscriber) Unsubscribe(strategyName string, codeList []string) (err error) {
	s.logger.Infof("unsubscribe strategy : %s code list : %q", strategyName, codeList)
	for _, code := range codeList {
		for i, name := range s.strategyMap[code] {
			if name == strategyName {
				s.strategyMap[code] = append(s.strategyMap[code][:i], s.strategyMap[code][i+1:]...)
			}
		}
	}
	return
}

func (s *Subscriber) getExternalIp() error {
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
	api.cache = make(map[string]*Quotation)
	if err != nil {
		return err
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("read message error: %s", err)
			return nil
		}
		raw := string(message)
		if strings.Contains(raw, "sys_auth=FAILED") {
			return fmt.Errorf("auth timeout")
		}
		rawLines := strings.SplitN(raw, "\n", -1)
		// @todo 如果有股票加入可能index的code会冲突
		for _, rawLine := range rawLines {
			if strings.Contains(rawLine, "sys_nxkey=") || strings.Contains(rawLine, "sys_time=") || len(rawLine) < 10 {
				continue
			}
			quo, err := api.parseQuotation(rawLine)
			if err == nil {
				if quo.TradeAmount == 0 && quo.Volume == 0 && quo.Close != quo.PreClose && quo.Close != 0 && quo.Bids[0].Price == quo.Asks[0].Price && quo.Bids[0].Amount == quo.Asks[0].Amount {
					quo.Code = "i" + quo.Code
				}
				api.subscriber.quotationCacheMap[quo.Code] = quo
				strategyNameList := api.strategyMap[quo.Code]
				for _, strategyName := range strategyNameList {
					api.quotationChanMap[strategyName] <- quo
				}
			}
		}
	}
}

func (api *Api) parseQuotation(rawLine string) (*Quotation, error) {
	quo := &Quotation{}
	rawLines := strings.SplitN(rawLine, "=", 2)
	if len(rawLines) < 2 {
		return quo, fmt.Errorf("unexpected data %s", rawLine)
	}
	if rawLines[0][0:4] == "2cn_" {
		quo.Code = rawLines[0][6:12]
		rawLines = strings.Split(rawLines[1], ",")
		length := len(rawLines)
		if length < 40 {
			return quo, fmt.Errorf("unexpected data %s", rawLine)
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

func (api *Api) refreshToken() error {
	log.Println("refresh token")
	client := &http.Client{}
	if api.Params[0:2] == "sh" || api.Params[0:2] == "sz" {
		api.Params = "2cn_sh502014," + api.Params
	}
	url := fmt.Sprintf(api.TokenServer, api.IP, api.Params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", api.UA)
	req.Header.Add("Cookie", api.Cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re := regexp.MustCompile(`result:"(.+?)"`)
	result := re.FindAllSubmatch(body, 1)
	if len(result) == 1 && len(result[0]) == 2 {
		api.token = string(result[0][1])
		log.Printf("get token %s", api.token)
		return nil
	} else {
		return fmt.Errorf("can't match token")
	}
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

func (s *Subscriber) GetQuation(code string) (quotation *Quotation, err error) {
	var found bool
	quotation, found = s.quotationCacheMap[code]
	if !found {
		err = fmt.Errorf("%s not coming", code)
		quotation = &Quotation{}
	}
	return
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
