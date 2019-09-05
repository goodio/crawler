package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const BOOK_STORE = `./book`

func main() {

	c := colly.NewCollector(
		colly.AllowedDomains("m.bqg5200.com"),
		//colly.Async(true),
	)

	extensions.RandomUserAgent(c)

	/*rp, err := proxy.RoundRobinProxySwitcher("http://103.108.47.17:30290","http://103.14.235.26:8080","http://103.93.237.74:3128",)

	if err != nil {
		logrus.Errorf("设置IP代理失败：%v", err)
	}

	c.SetProxyFunc(rp)*/

	//var reg = regexp.MustCompile(`\/\w+-\d+-(\d+)`)
	var reg = regexp.MustCompile(`https:\/\/m.bqg5200.com\/wapbook-\d+-(\d+)*`)

	c.OnHTML("body#nr_body", func(e *colly.HTMLElement) {

		upath := e.Request.URL.String()

		fname := reg.FindStringSubmatch(upath)

		h, _ := e.DOM.Html()

		html, err := DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		book_name := dom.Find("h1#_52mb_h1").Text()

		title := strings.TrimSpace(dom.Find("div#nr_title").Text())

		article := strings.Replace(dom.Find("div#nr1").Text(), "聽", " ", -1)

		content := "### " + title + "\n" + article + "\n\n"

		filepath := path.Join(book_name, fname[1])

		err = write(filepath, content)

		if err != nil {
			fmt.Printf("%v\n", err)
		}

	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		c.Visit(link)

		/*if reg.MatchString(link) {
			c.Visit(link)
		}*/
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "m.bqg5200.com/*",
		RandomDelay: 2 * time.Second,
		Parallelism: 5,
	})

	c.OnRequest(func(r *colly.Request) {
		//time.Sleep(getRandomDelay(1000))
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://m.bqg5200.com")
	//c.Visit("https://m.bqg5200.com/wapbook-24282-10406931/")

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
