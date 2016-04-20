package gotrade

// import (
// 	"fmt"
// 	"github.com/Sirupsen/logrus"
// 	"github.com/mreiferson/go-httpclient"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"net/http/cookiejar"
// 	"strings"
// 	"time"
// )

// type Xueqiu struct {
// 	client        *http.Client   `yaml:"-"`
// 	Account       string         `yaml:"account"`
// 	Password      string         `yaml:"password"`
// 	PortfolioCode string         `yaml:"portfolio_code"`
// 	logger        *logrus.Logger `yaml:"-"`
// }

// func NewXueqiu(configPath string) (x *Xueqiu) {
// 	x = &Xueqiu{}
// 	x.logger = NewLogger("xueqiu")
// 	log.Println(GetBasePath() + configPath)
// 	err := YamlFileDecode(GetBasePath()+configPath, &x)
// 	log.Println(x)
// 	if err != nil {
// 		x.logger.Error("init account error")
// 		panic(err)
// 	}
// 	return
// }

// func (x *Xueqiu) Login() (err error) {
// 	cookieJar, _ := cookiejar.New(nil)
// 	transport := &httpclient.Transport{
// 		ConnectTimeout:        3 * time.Second,
// 		RequestTimeout:        3 * time.Second,
// 		ResponseHeaderTimeout: 3 * time.Second,
// 	}
// 	defer transport.Close()
// 	x.client = &http.Client{
// 		CheckRedirect: nil,
// 		Jar:           cookieJar,
// 		Transport:     transport,
// 	}
// 	x.logger.Info("begin login xueqiu")
// 	raw := fmt.Sprintf("username=&areacode=86&telephone=%s&remember_me=1&password=%s", x.Account, x.Password)
// 	req := x.newRequest("POST", "http://xueqiu.com/user/login", raw)
// 	resp, _ := x.client.Do(req)
// 	body, _ := ioutil.ReadAll(resp.Body)
// 	type Result struct {
// 		Uid string `json:"uid"`
// 	}
// 	result := Result{}
// 	json.Unmarshal([]byte(body), &result)
// 	if result.Uid == "" {
// 		return fmt.Errorf("format", string(body))
// 	}

// 	log.Println(string(body))
// 	return
// }

// func (x *Xueqiu) newRequest(method string, url string, raw string) (req *http.Request) {
// 	req, _ = http.NewRequest(method, url, strings.NewReader(raw))
// 	log.Println(method, url, raw)
// 	req.Header.Add("Origin", "xueqiu.com")
// 	req.Header.Add("Pragma", "no-cache")
// 	req.Header.Add("cache-control", "no-cache")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
// 	req.Header.Add("Connection", "Connection")
// 	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6,zh-TW;q=0.4,ja;q=0.2")
// 	req.Header.Add("X-Requested-With", "XMLHttpRequest")
// 	req.Header.Add("Referer", "http://xueqiu.com/8864209811")
// 	req.Header.Add("Accept", "*/*")
// 	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36")
// 	return req
// }
