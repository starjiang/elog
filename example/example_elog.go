package main

import (
	"flag"
	"time"

	"github.com/starjiang/elog"
)

func main() {
	flag.Parse()
	defer elog.Flush()
	for i := 0; i < 100; i++ {
		go func() {
			elog.Info("hello", "world")
		}()
	}
	time.Sleep(time.Second)
}
