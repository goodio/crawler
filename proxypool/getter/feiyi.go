package getter

import (
	"github.com/Aiicy/htmlquery"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

//feiyi get ip from feiyiproxy.com
func Feiyi() (result []*proxypool.IP) {
	logrus.Info("FEIYI] start test")
	pollURL := "http://www.feiyiproxy.com/?page_id=1457"
	doc, _ := htmlquery.LoadURL(pollURL)
	trNode, err := htmlquery.Find(doc, "//div[@class='et_pb_code.et_pb_module.et_pb_code_1']//div//table//tbody//tr")
	logrus.Info("[FEIYI] start up")
	if err != nil {
		logrus.Info("FEIYI] parse pollUrl error")
		logrus.Warn(err.Error())
	}
	//debug begin
	logrus.Info("[FEIYI] len(trNode) = %d ", len(trNode))
	for i := 1; i < len(trNode); i++ {
		tdNode, _ := htmlquery.Find(trNode[i], "//td")
		ip := htmlquery.InnerText(tdNode[0])
		port := htmlquery.InnerText(tdNode[1])
		Type := htmlquery.InnerText(tdNode[3])
		speed := htmlquery.InnerText(tdNode[6])

		IP := &proxypool.IP{}
		IP.Data = ip + ":" + port

		if Type == "HTTPS" {
			IP.Type1 = "https"
			IP.Type2 = ""

		} else if Type == "HTTP" {
			IP.Type1 = "http"
		}
		IP.Speed = extractSpeed(speed)

		logrus.Info("[FEIYI] ip.Data = %s,ip.Type = %s,%s ip.Speed = %d", IP.Data, IP.Type1, IP.Type2, IP.Speed)

		result = append(result, IP)
	}

	logrus.Info("FEIYI done.")
	return
}

func extractSpeed(oritext string) int64 {
	reg := regexp.MustCompile(`\[1-9\]\d\*\\.\?\d\*`)
	temp := reg.FindString(oritext)
	if temp != "" {
		speed, _ := strconv.ParseInt(temp, 10, 64)
		return speed
	}
	return -1
}
