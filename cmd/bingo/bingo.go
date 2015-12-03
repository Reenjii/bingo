package main

import (
	"flag"
	"log"

	"github.com/reenjii/bingo"
)

var conf string

func init() {
	//log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	bingo.Loggers.Info.Println("Bingo initialization")

	flag.StringVar(&conf, "conf", "/etc/bingo.json", "Configuration file path")
}

func main() {
	flag.Parse()
	bingo.Serve(conf)
}
