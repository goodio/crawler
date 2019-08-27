package initial

import (
	"github.com/ghaoo/crawler/proxypool"
	"github.com/ghaoo/crawler/proxypool/getter"
	"github.com/sirupsen/logrus"
	"sync"
)

func Run(ipChan chan<- *proxypool.IP) {
	logrus.Info("初始化IP代理池...")
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
		getter.PLP, //need to remove it
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
	logrus.Info("IP代理池初始化完成...")
}
