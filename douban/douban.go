package main

import (
	"github.com/sirupsen/logrus"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"github.com/ghaoo/crawler/proxypool"
	"time"
	"math/rand"
)

func main() {

	go proxypool.Go()

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.AllowedDomains("book.douban.com/subject"),
	)

	proxyIp := proxypool.RondomIP()
	rp, err := proxy.RoundRobinProxySwitcher(proxyIp.Type1 + "://" + proxyIp.Data)

	if err != nil {
		logrus.Errorf("设置IP代理失败：%v", err)
	}

	c.SetProxyFunc(rp)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		c.Visit(e.Request.AbsoluteURL(link))
	})

	douban(c)

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(getRandomDelay())

		logrus.Debugf("爬取地址：%s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		//logrus.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

	c.Visit("https://book.douban.com/")
}

func douban(c *colly.Collector) {
	c.OnHTML("#wrapper", func(e *colly.HTMLElement) {
		book_name := e.DOM.Find("h1").First().Find("span").Text()

		logrus.Info("书名: ", book_name)
	})

}

// 随机延时
func getRandomDelay() time.Duration {
	return time.Duration(rand.Int63n(2000)) * time.Millisecond
}


func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}
