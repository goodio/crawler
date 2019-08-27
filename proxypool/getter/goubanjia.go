package getter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// GBJ get ip from goubanjia.com
func GBJ() (result []*proxypool.IP) {
	pollURL := "http://www.goubanjia.com/"

	resp, _, errs := gorequest.New().Get(pollURL).End()
	if errs != nil {
		logrus.Error(errs)
		return
	}
	fmt.Println(resp.Body)
	if resp.StatusCode != 200 {
		logrus.Println(resp.StatusCode)
		logrus.Warn(errs)
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	doc.Find("body > div.container > div.section-header > div.row > div.container-fluid > div.row-fluid > div.span12 > tbody > tr").Each(func(_ int, s *goquery.Selection) {
		sf, _ := s.Find(".ip").Html()
		tee := regexp.MustCompile("<pstyle=\"display:none;\">.?.?</p>").ReplaceAllString(strings.Replace(sf, " ", "", -1), "")
		re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
		ip := &proxypool.IP{}
		ip.Data = re.ReplaceAllString(tee, "")
		ip.Type1 = s.Find("td:nth-child(3) > a").Text()
		//logrus.Printf("ip.Data = %s , ip.Type = %s\n", ip.Data, ip.Type1)
		result = append(result, ip)
	})

	//logrus.Info("GBJ done.")
	return
}
