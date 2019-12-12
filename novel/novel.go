package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
	"strings"
	"strconv"
	"net"
	"github.com/ghaoo/rbootx/tools"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"encoding/json"
	"path/filepath"
	"bufio"
	"io"
)

var url = `https://www.bqg5200.com/xiaoshuo/23/23730/`

type Catalog struct {
	ID          int       // ID
	SubID       string    // SUB ID
	Name        string    // 名称
	Author      string    // 作者
	Url         string    // 链接
	Chapters    []Chapter // 章节ID列表
	Category    string    // 类别
	LastChapter string    // 最新章节
	LastUpdate  string    // 最后更新
}

type Chapter struct {
	ID   int
	Url  string
	Name string
}

func main() {
	cat := GetCatalog(url)

	fetchContent(&cat)

	path := cat.Name

	fileMerge(path)
}

func GetCatalog(url string) Catalog {
	cl := Catalog{}

	logrus.Warn(url)

	c := colly.NewCollector(
		colly.AllowedDomains("www.bqg5200.com"),
	)

	c.Limit(&colly.LimitRule{
		Parallelism: 1,
		RandomDelay: 5 * time.Second,
	})

	c.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	extensions.RandomUserAgent(c)

	var reg = regexp.MustCompile(`https:\/\/www.bqg5200.com\/xiaoshuo\/(\d+)\/(\d+)[\/]?$`)
	var reg2 = regexp.MustCompile(`https:\/\/www.bqg5200.com\/xiaoshuo\/\d+\/\d+\/(\d+).html`)
	c.OnHTML("div#maininfo", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()

		idr := reg.FindStringSubmatch(url)

		subid := idr[1]

		idstr := idr[2]

		id, _ := strconv.Atoi(idstr)

		h, _ := e.DOM.Html()

		html, _ := tools.DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		title := dom.Find("div.coverecom div:nth-of-type(2)")

		name := title.Find("h1").Text()

		author := title.Find("span:first-of-type").Text()

		category := title.Find("span:nth-of-type(2) a").Text()

		last_update := title.Find("span:nth-of-type(3)").Text()

		last_chapter := dom.Find("#readerlist ul li:last-of-type a").Text()

		cpts := []Chapter{}
		dom.Find("#readerlist ul li").Each(func(i int, s *goquery.Selection) {

			cname := s.Find("a").Text()
			curl, _ := s.Find("a").Attr("href")
			curl = e.Request.AbsoluteURL(curl)

			if reg2.MatchString(curl) {
				cid, err := strconv.Atoi(reg2.FindStringSubmatch(curl)[1])

				if err != nil {
					cid = 0
				}

				cpt := Chapter{
					ID:   cid,
					Name: cname,
					Url:  curl,
				}

				cpts = append(cpts, cpt)
			}

		})

		cl.ID = id
		cl.SubID = subid
		cl.Name = name
		cl.Author = author
		cl.Url = url
		cl.Category = category
		cl.Chapters = cpts
		cl.LastChapter = last_chapter
		cl.LastUpdate = last_update

		fname := path.Join(os.Getenv("BOOK_PATH"), name, "data.json")

		data, err := json.Marshal(&cl)
		if err != nil {
			logrus.Error(err)
		} else {
			tools.FileWrite(fname, data)
		}

	})

	c.Visit(url)

	c.Wait()

	return cl
}

func fetchContent(cl *Catalog) {

	c := colly.NewCollector(
		colly.AllowedDomains("www.bqg5200.com"),
		//colly.DisallowedURLFilters(regexp.MustCompile(`https:\/\/m.bqg5200.com\/wapbook-753-(\d+)*`)),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainRegexp: "www.bqg5200.com/*",
		Parallelism:  30,
		RandomDelay:  5 * time.Second,
	})

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	extensions.RandomUserAgent(c)

	var reg = regexp.MustCompile(`https:\/\/www.bqg5200.com\/xiaoshuo\/\d+\/\d+\/(\d+).html`)

	c.OnHTML("body.clo_bg", func(e *colly.HTMLElement) {

		upath := e.Request.URL.String()

		fname := reg.FindStringSubmatch(upath)

		h, _ := e.DOM.Html()

		html, _ := tools.DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		book_name := dom.Find("#header .readNav :nth-child(3)").Text()

		title := strings.TrimSpace(dom.Find("div.title h1").Text())

		dom.Find("div#content div").Remove()
		article, _ := dom.Find("div#content").Html()
		article = strings.Replace(article, "聽", " ", -1)
		article = strings.Replace(article, "<br/>", "\n", -1)

		content := "### " + title + "\n" + article + "\n\n"

		fpath := filepath.Join(book_name, fname[1]+".rbx")

		err := tools.FileWrite(fpath, []byte(content))

		if err != nil {
			logrus.Errorf("%v\n", err)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second)
		logrus.Debugf("Visiting %s", r.URL.String())
	})

	for _, cpt := range cl.Chapters {

		c.Visit(cpt.Url)
	}

	c.Wait()

}

func fileMerge(root string) error {
	name := filepath.Base(root)

	out_name := filepath.Join(root, name+".txt")

	out_file, err := os.OpenFile(out_name, os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		return fmt.Errorf("Can not open file %s", out_name)
	}

	bWriter := bufio.NewWriter(out_file)

	bWriter.Write([]byte("## " + name + "\n\n\n"))

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && strings.HasSuffix(path, ".rbx") {
			//logrus.Printf("读取文件：%s \n", info.Name())

			fp, err := os.Open(path)

			if err != nil {
				fmt.Printf("Can not open file %v", err)
				return err
			}

			defer fp.Close()

			bReader := bufio.NewReader(fp)

			for {

				buffer := make([]byte, 1024)
				readCount, err := bReader.Read(buffer)
				if err == io.EOF {
					break
				} else {
					bWriter.Write(buffer[:readCount])
				}

			}

			bWriter.Write([]byte("\n\n"))
		}

		return err
	})

	bWriter.Flush()

	return nil
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}

