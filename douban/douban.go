package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type Book struct {
}

func main() {

	c := colly.NewCollector(
		//colly.AllowURLRevisit(),
		colly.AllowedDomains("book.douban.com"),
	)

	/*c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})*/

	extensions.RandomUserAgent(c)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnHTML("#wrapper", func(e *colly.HTMLElement) {
		book_name := e.DOM.Find("h1").First().Find("span").Text()

		logrus.Info("书名: ", book_name)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(getRandomDelay() + 5*time.Second)

		logrus.Infof("爬取地址：%s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		//logrus.Info("代理IP: ", r.Request.ProxyURL)
	})

	c.Visit("https://book.douban.com/subject/24845582/")
	//c.Wait()
}

// 随机延时
func getRandomDelay() time.Duration {
	return time.Duration(rand.Int63n(8000)) * time.Millisecond
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}
