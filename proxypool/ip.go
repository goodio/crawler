package proxypool

import (
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"time"
	"encoding/json"
	"sync"
	"github.com/ghaoo/crawler/proxypool/getter"
)

const BUCKET_NAME = `proxypool`

var db = NewBolt()

// IP struct
type IP struct {
	Data  string
	Type1 string
	Type2 string
	Speed int64
}

func Getter() {
	ipChan := make(chan *IP, 2000)

	go func() {
		CheckProxyDB()
	}()

	for i := 0; i < 50; i++ {
		go func() {
			for {
				CheckAndSave(<-ipChan)
			}
		}()
	}

	for {
		if len(ipChan) < 100 {
			go run(ipChan)
		}
		time.Sleep(10 * time.Minute)
	}
}

func run(ipChan chan<- *IP) {
	var wg sync.WaitGroup
	funs := []func() []*IP{
		getter.Feiyi,
		getter.IP66, //need to remove it
		getter.KDL,
		//getter.GBJ,	//因为网站限制，无法正常下载数据
		//getter.Xici,
		//getter.XDL,
		//getter.IP181,  // 已经无法使用
		//getter.YDL,	//失效的采集脚本，用作系统容错实验
		getter.PLP,   //need to remove it
		getter.IP89,
	}
	for _, f := range funs {
		wg.Add(1)
		go func(f func() []*IP) {
			temp := f()

			for _, v := range temp {

				ipChan <- v
			}
			wg.Done()
		}(f)
	}
	wg.Wait()
	logrus.Println("All getters finished.")
}

func CheckAndSave(ip *IP) {
	if CheckIP(ip) {
		ipjs, _ := json.Marshal(ip)
		err := db.Save(BUCKET_NAME, ip.Data, ipjs)

		if err != nil {
			logrus.Error("[CheckAndSave] Error = %v", err)
		}
	}
}

func CheckProxyDB() {

	ips := db.FindAll(BUCKET_NAME)
	if len(ips) <= 0 {
		logrus.Warn("not found")
		return
	}
	var wg sync.WaitGroup
	for _, v := range ips {

		i := &IP{}

		if err := json.Unmarshal(v, &i); err != nil {
			logrus.Errorf("Unmarshal IP Error = %v", err)
		}

		wg.Add(1)
		go func(ip *IP) {
			if !CheckIP(ip) {
				Delete(ip)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func Delete(ip *IP) {
	err := db.Delete(BUCKET_NAME, ip.Data)

	if err != nil {
		logrus.Error("[delete] Error = %v", err)
	}
}

func CheckIP(ip *IP) bool {
	var pollURL string
	var testIP string
	if ip.Type2 == "https" {
		testIP = "https://" + ip.Data
		pollURL = "https://httpbin.org/get"
	} else {
		testIP = "http://" + ip.Data
		pollURL = "http://httpbin.org/get"
	}
	logrus.Warningf(testIP)
	begin := time.Now()
	resp, _, errs := gorequest.New().Proxy(testIP).Get(pollURL).End()
	if errs != nil {
		logrus.Warningf("[CheckIP] testIP = %s, pollURL = %s: Error = %v", testIP, pollURL, errs)
		return false
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		//harrybi 计算该代理的速度，单位毫秒
		ip.Speed = time.Now().Sub(begin).Nanoseconds() / 1000 / 1000 //ms

		ipjs, _ := json.Marshal(ip)
		if err := db.Update(BUCKET_NAME, ip.Data, ipjs); err != nil {
			logrus.Warningf("[CheckIP] Update IP = %v Error = %v", *ip, err)
		}

		return true
	}
	return false
}
