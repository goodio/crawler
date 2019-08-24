package getter

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"strings"
)

// YDL get ip from youdaili.net
func YDL() (result []*proxypool.IP) {
	pollURL := "http://www.youdaili.net/Daili/http/"
	_, body, errs := gorequest.New().Get(pollURL).End()
	if errs != nil {
		logrus.Error(errs)
		return
	}
	do, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		logrus.Warn(err.Error())
		return
	}

	URL, _ := do.Find("body > div.con.PT20 > div.conl > div.lbtc.l > div.chunlist > ul > li:nth-child(1) > p > a").Attr("href")
	_, content, errs := gorequest.New().Get(URL).End()
	if errs != nil {
		logrus.Error(errs)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	doc.Find(".content p").Each(func(_ int, s *goquery.Selection) {
		ip := &proxypool.IP{}
		c := strings.Split(s.Text(), "@")
		ip.Data = c[0]
		ip.Type1 = strings.ToLower(strings.Split(c[1], "#")[0])
		result = append(result, ip)
	})
	logrus.Println("YDL done.")
	return
}
