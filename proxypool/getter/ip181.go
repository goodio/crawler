package getter

import (
	"encoding/json"
	"io/ioutil"

	"github.com/parnurzeal/gorequest"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/sirupsen/logrus"
)

type ip181 struct {
	ErrorCode string   `json:"ERRORCODE"`
	Results   []Result `json:"RESULT"`
}

// Result struct
type Result struct {
	Position string `json:"position"`
	Port     string `json:"port"`
	IP       string `json:"ip"`
}

// IP181 get ip from ip181.com
func IP181() (result []*proxypool.IP) {
	var ips ip181
	var results []Result

	pollURL := "http://www.ip181.com/"
	resp, _, errs := gorequest.New().Get(pollURL).End()
	if errs != nil {
		logrus.Error(errs)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	err := json.Unmarshal(body, &ips)

	if err != nil {
		logrus.Error(err)
	}

	results = ips.Results

	for i := 0; i < len(results); i++ {
		ip := &proxypool.IP{}
		ip.Data = results[i].IP + ":" + results[i].Port
		ip.Type1 = "http"
		logrus.Info("[IP181] ip.Data: %s,ip.Type: %s", ip.Data, ip.Type1)
		result = append(result, ip)
	}

	logrus.Info("IP181 done.")
	return
}
