package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ProxyIp struct {
	Ip                      string
	Port                    int
	IsHttps                 bool
	UpdateTime              int
	SourceUrl               string
	TimeTolive              int
	AnonymousInfo           string
	Area                    string
	InternetServiceProvider string
}

var ProxyIpPool []ProxyIp

func main() {
	p := &ProxyIpPool
	SourceUrl := "http://www.xicidaili.com/wt/"
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Limit the maximum parallelism to 1
	// This is necessary if the goroutines are dynamically
	// created to control the limit of simultaneous requests.
	//
	// Parallelism can be controlled also by spawning fixed
	// number of go routines.
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 12})

	// On every a element which has href attribute call callback
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var item ProxyIp
		e.ForEach("td", func(i int, element *colly.HTMLElement) {
			t := element.Text
			switch i {
			case 1:
				item.Ip = t
				break
			case 2:
				p, n := strconv.Atoi(t)
				if n == nil {
					item.Port = p
				}
				break
			case 3:
				item.Area = t
				break
			case 4:
				item.IsHttps = strings.Contains(strings.ToLower(t), "https")
				break
			default:
				break
			}

		})
		item.SourceUrl = SourceUrl
		*p = append(*p, item)
	})

	// Start scraping on https://en.wikipedia.org
	c.Visit(SourceUrl)
	// Wait until threads are finished
	c.Wait()

	//var a [] string
	for _, v := range *p {
		http := "http"
		if v.IsHttps {
			http = "https"
		}
		if v.Ip != "" && v.Port != 0 {
			s := http + "://" + v.Ip + ":" + strconv.Itoa(v.Port)
			ip, status := ProxyThorn(s)

			if status == 200 && ip != "" {
				logrus.Println(s + " 请求 http://icanhazip.com 返回ip:【" + ip + "】-【检测结果：可用】")
			} else {
				logrus.Println(s + " 请求 http://icanhazip.com 返回ip:【" + ip + "】-【检测结果：不可用】")
			}

			//fmt.Println(s)
			//a = append(a, s)
		}
	}

	// Instantiate default collector
	c = colly.NewCollector(colly.AllowURLRevisit(), colly.Async(true))

	// Rotate two socks5 proxies
	//rp, err := proxy.RoundRobinProxySwitcher("http://113.124.92.182:9999", "https://119.254.94.114:45691", "https://124.232.133.199:3128", "http://182.88.11.236:9797")
	rp, err := proxy.RoundRobinProxySwitcher( /*"http://182.88.11.236:9797", */ "http://111.160.121.238:8080")
	if err != nil {
		logrus.Fatal(err)
	}
	c.SetProxyFunc(rp)

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		logrus.Printf("Proxy Address: %s\n", r.Request.ProxyURL)
		//log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

	// Fetch httpbin.org/ip five times
	for i := 0; i < 15; i++ {
		//c.Visit("https://httpbin.org/ip")
		c.Visit("http://ad.dxinsw.net")
	}

	c.Wait()

}

func ProxyThorn(proxy_addr string) (ip string, status int) {
	//访问查看ip的一个网址
	httpUrl := "http://icanhazip.com"
	proxy, err := url.Parse(proxy_addr)

	netTransport := &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxIdleConnsPerHost:   10,
		TLSHandshakeTimeout:   1 * time.Second,
		ResponseHeaderTimeout: time.Second * time.Duration(1),
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	res, err := httpClient.Get(httpUrl)
	if err != nil {
		//fmt.Println("错误信息：",err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logrus.Println(err)
		return
	}
	c, _ := ioutil.ReadAll(res.Body)
	return string(c), res.StatusCode
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}
