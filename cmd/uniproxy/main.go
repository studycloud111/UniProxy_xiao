package main

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/wyx2685/UniProxy/conf"
	"github.com/wyx2685/UniProxy/proxy"
	"github.com/wyx2685/UniProxy/router"
)

var host = flag.String("host", "127.0.0.1", "host")
var port = flag.Int("port", 11451, "port")
var config = flag.String("conf", "", "config file")
var tun = flag.Bool("tun", false, "tun mode")

func main() {
	flag.Parse()
	err := conf.Init(*config)
	if err != nil {
		log.WithField("err", err).Fatalln("init conf failed")
	}
	switch conf.C.Log.Level {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
	proxy.TunMode = *tun
	proxy.ResUrl = "http://127.0.0.1:" + strconv.Itoa(*port)
	router.Init()
	if err := router.Start(*host, *port); err != nil {
		log.Fatalln("start error:", err)
	}
}
