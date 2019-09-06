package main

import (
	"bytes"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
	"strings"
)

const BOOK_STORE = `E:\data\book`
//const BOOK_STORE = `./book`

func main() {

	c := colly.NewCollector(
		colly.AllowedDomains("www.bqg5200.com"),
		//colly.DisallowedURLFilters(regexp.MustCompile(`https:\/\/m.bqg5200.com\/wapbook-753-(\d+)*`)),
		colly.Async(true),
	)

	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			//KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	extensions.RandomUserAgent(c)

	var reg = regexp.MustCompile(`https:\/\/www.bqg5200.com\/xiaoshuo\/\d+\/\d+\/(\d+).html`)

	c.OnHTML("body.clo_bg", func(e *colly.HTMLElement) {

		upath := e.Request.URL.String()

		fname := reg.FindStringSubmatch(upath)

		h, _ := e.DOM.Html()

		html, _ := DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		class_name := dom.Find("#header .readNav :nth-child(2)").Text()

		book_name := dom.Find("#header .readNav :nth-child(3)").Text()

		title := strings.TrimSpace(dom.Find("div.title h1").Text())

		dom.Find("div#content div").Remove()
		article, _ := dom.Find("div#content").Html()
		article = strings.Replace(article, "聽", " ", -1)
		article = strings.Replace(article, "<br/>", "\n", -1)

		content := "### " + title + "\n" + article + "\n\n"

		filepath := path.Join(class_name, book_name, fname[1])

		err := write(filepath, content)

		if err != nil {
			logrus.Errorf("%v\n", err)
		}

	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		//c.Visit(link)

		if strings.HasPrefix(link, `https://www.bqg5200.com/book`) || strings.HasPrefix(link, `https://www.bqg5200.com/xiaoshuo`) {
			c.Visit(link)
		}
	})

	c.Limit(&colly.LimitRule{
		RandomDelay:  2 * time.Second,
		Parallelism:  5,
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(getRandomDelay(10000))
		logrus.Infof("Visiting %s", r.URL.String())
	})

	c.Visit("https://www.bqg5200.com")

	c.Wait()

}

func write(file, content string) error {

	filepath := path.Join(BOOK_STORE, file)

	basepath := path.Dir(filepath)
	// 检测文件夹是否存在   若不存在  创建文件夹
	if _, err := os.Stat(basepath); err != nil {

		if os.IsNotExist(err) {

			err = os.MkdirAll(basepath, os.ModePerm)

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	data := []byte(content)

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
}

func getRandomDelay(seed int64) time.Duration {
	return time.Duration(rand.Int63n(seed+1000)) * time.Millisecond
}

func DecodeGBK(s []byte) ([]byte, error) {
	reader := simplifiedchinese.GB18030.NewDecoder().Reader(bytes.NewReader(s))

	return ioutil.ReadAll(reader)
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}

