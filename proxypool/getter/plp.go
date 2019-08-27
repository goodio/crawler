package getter

import (
	"github.com/Aiicy/htmlquery"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/sirupsen/logrus"
)

//PLP get ip from proxylistplus.com
func PLP() (result []*proxypool.IP) {
	pollURL := "https://list.proxylistplus.com/Fresh-HTTP-Proxy-List-1"
	doc, _ := htmlquery.LoadURL(pollURL)
	trNode, err := htmlquery.Find(doc, "//div[@class='hfeed site']//table[@class='bg']//tbody//tr")
	if err != nil {
		logrus.Warn(err.Error())
	}
	for i := 3; i < len(trNode); i++ {
		tdNode, _ := htmlquery.Find(trNode[i], "//td")
		ip := htmlquery.InnerText(tdNode[1])
		port := htmlquery.InnerText(tdNode[2])
		Type := htmlquery.InnerText(tdNode[6])

		IP := &proxypool.IP{}
		IP.Data = ip + ":" + port

		if Type == "yes" {
			IP.Type1 = "http"
			IP.Type2 = "https"

		} else if Type == "no" {
			IP.Type1 = "http"
		}

		//logrus.Infof("[PLP] ip.Data = %s,ip.Type = %s,%s", IP.Data, IP.Type1, IP.Type2)

		result = append(result, IP)
	}

	//logrus.Info("PLP done.")
	return
}
