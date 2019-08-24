package main

import (
	"github.com/ghaoo/crawler/proxypool"
	"github.com/sirupsen/logrus"
)

func main() {

	proxypool.Go()

	proxypool.RondomIP()
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}
