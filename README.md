# elog
package main

import (
	"flag"
	"time"

	"github.com/starjiang/elog"
)

func main() {
	flag.Parse()
	defer elog.Flush()
	elog.Info("hello","world")
  elog.Infof("hello %s","world")  
}

./main -logToStderr -logLevel=INFO -logFlushTime=3 -logPath=./
