package getter

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/ghaoo/crawler/proxypool"
	"github.com/sirupsen/logrus"
)

// IP66 get ip from 66ip.cn
func IP66() (result []*proxypool.IP) {
	var ExprIP = regexp.MustCompile(`((25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))\.){3}(25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))\:([0-9]+)`)

	pollURL := "http://www.66ip.cn/mo.php?tqsl=100"
	resp, err := http.Get(pollURL)
	if err != nil {
		logrus.Error(err)
		return
	}

	if resp.StatusCode != 200 {
		logrus.Warn(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodyIPs := string(body)
	ips := ExprIP.FindAllString(bodyIPs, 100)

	for index := 0; index < len(ips); index++ {
		ip := &proxypool.IP{}
		ip.Data = strings.TrimSpace(ips[index])
		ip.Type1 = "http"
		//logrus.Infof("[IP66] ip = %s, type = %s", ip.Data, ip.Type1)
		result = append(result, ip)
	}

	//logrus.Info("IP66 done.")
	return
}
