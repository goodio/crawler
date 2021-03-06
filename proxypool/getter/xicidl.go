package getter

import (
	"github.com/ghaoo/crawler/proxypool"
	"github.com/nladuo/go-phantomjs-fetcher"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

// Xici get ip from xicidaili.com
func Xici() (result []*proxypool.IP) {
	pollURL := "http://www.xicidaili.com/nn/"

	fetcher, err := phantomjs.NewFetcher(2017, nil)
	defer fetcher.ShutDownPhantomJSServer()
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	jsScript := "function() {s=document.documentElement.outerHTML;document.write('<body></body>');document.body.innerText=s;}"
	jsRunAt := phantomjs.RUN_AT_DOC_END
	resp, err := fetcher.GetWithJS(pollURL, jsScript, jsRunAt)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	re, _ := regexp.Compile("<td>(\\d+\\.){3}\\d+</td>.+?(\\d{2,4})</td>")
	temp := re.FindAllString(strings.Replace(strings.Replace(resp.Content, "&lt;", "<", -1), "&gt;", ">", -1), -1)

	for _, v := range temp {
		v = strings.Replace(v, "<td>", "", -1)
		v = strings.Replace(v, "</td>", "", -1)
		v = strings.Replace(v, " ", "", -1)
		v = strings.Replace(v, "<br>", ":", -1)
		ip := &proxypool.IP{}
		ip.Data = v
		ip.Type1 = "http"
		result = append(result, ip)
	}
	//logrus.Println("Xici done.")
	return
}
