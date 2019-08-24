package pool

import (
	"time"
	"sync"
	"github.com/sirupsen/logrus"
	"github.com/ghaoo/crawler/proxypool"
	"github.com/ghaoo/crawler/proxypool/getter"
)

func Go() {
	ipChan := make(chan *proxypool.IP, 2000)

	go func() {
		proxypool.CheckProxyDB()
	}()

	for i := 0; i < 50; i++ {
		go func() {
			for {
				proxypool.CheckAndSave(<-ipChan)
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

func run(ipChan chan<- *proxypool.IP) {
	var wg sync.WaitGroup
	funs := []func() []*proxypool.IP{
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
		go func(f func() []*proxypool.IP) {
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
