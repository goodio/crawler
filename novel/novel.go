package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
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
	"io/ioutil"
	"sort"
	"io"
	"path"
)

//var url = `https://www.bqg5200.com/xiaoshuo/23/23730/`
var url = `https://www.cnoz.org/0_1/`

type Catalog struct {
	ID          string    // ID
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
	cat := CNOZ_GetCatalog(url)

	fetchContent(cat)

	path := cat.Name

	fileMerge(path)
}

func CNOZ_GetCatalog(url string) *Catalog {
	cl := &Catalog{}

	logrus.Info(url)

	c := colly.NewCollector(
		colly.AllowedDomains("www.cnoz.org"),
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

	var reg = regexp.MustCompile(`https:\/\/www.cnoz.org\/0_(\d+)[\/]?$`)
	var reg2 = regexp.MustCompile(`https:\/\/www.cnoz.org\/0_\d+\/(\d+).html`)

	c.OnHTML("div#wrapper", func(e *colly.HTMLElement) {

		url := e.Request.URL.String()

		idr := reg.FindStringSubmatch(url)

		id := idr[1]

		h, _ := e.DOM.Html()

		html, _ := tools.DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		title := dom.Find("div#maininfo div#info")

		name := title.Find("h1").Text()

		var cpts = make([]Chapter, 0)
		dom.Find("div#list dl dd").Each(func(i int, s *goquery.Selection) {

			if i >= 9 {
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
			}

		})

		cl.ID = id
		cl.Name = name
		cl.Url = url
		cl.Chapters = cpts

		fname := path.Join(name, "data.json")
		fmt.Println(fname)

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

	c := colly.NewCollector()

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	extensions.RandomUserAgent(c)

	var reg = regexp.MustCompile(`https:\/\/www.cnoz.org\/0_\d+\/(\d+).html`)

	c.OnHTML(".content_read div.box_con", func(e *colly.HTMLElement) {

		upath := e.Request.URL.String()

		fname := reg.FindStringSubmatch(upath)

		h, _ := e.DOM.Html()

		html, _ := tools.DecodeGBK([]byte(h))

		dom := e.DOM.SetHtml(string(html))

		title := dom.Find(".bookname h1").Text()

		dom.Find("div#content div").Remove()
		article, _ := dom.Find("div#content").Html()
		article = strings.Replace(article, "聽", " ", -1)
		article = strings.Replace(article, "<br/>", "\n", -1)

		content := "### " + title + "\n" + article + "\n\n"

		fpath := path.Join(cl.Name, fname[1]+".rbx")

		err := tools.FileWrite(fpath, []byte(content))

		if err != nil {
			logrus.Errorf("%v\n", err)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second)
		logrus.Infof("访问 %s", r.URL.String())
	})

	for _, cpt := range cl.Chapters {

		c.Visit(cpt.Url)
	}

	c.Wait()

}

func fileMerge(root string) error {
	name := filepath.Base(root)

	out_name := path.Join(root, name+".txt")

	out_file, err := os.OpenFile(out_name, os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		return fmt.Errorf("Can not open file %s", out_name)
	}

	bWriter := bufio.NewWriter(out_file)

	bWriter.Write([]byte("## " + name + "\n\n\n"))

	files, _ := ioutil.ReadDir(root)

	sort.SliceStable(files, func(i, j int) bool {
		name1 := strings.TrimSuffix(files[i].Name(), ".rbx")
		name2 := strings.TrimSuffix(files[j].Name(), ".rbx")

		f1, _ := strconv.Atoi(name1)
		f2, _ := strconv.Atoi(name2)

		return  f1 < f2
	})

	for _, file := range files{
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".rbx") {
			logrus.Printf("读取文件：%s \n", file.Name())

			fp, err := os.Open(path.Join(root, file.Name()))

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

			bWriter.Write([]byte("\n"))
		}
	}

	/*filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && strings.HasSuffix(path, ".rbx") {
			logrus.Printf("读取文件：%s \n", info.Name())

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

			bWriter.Write([]byte("\n"))
		}

		return err
	})*/

	bWriter.Flush()

	return nil
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}

