package main

import (
	"os"

	"github.com/starjiang/elog"
)

type TestLogHandler struct {
}

func (lh *TestLogHandler) Write(data []byte) (int, error) {

	return os.Stderr.Write(data)
}

func (lh *TestLogHandler) Flush() {
	os.Stderr.Write([]byte("flush\n"))
}

func main() {
	writer := &TestLogHandler{}
	log := elog.NewEasyLogger("INFO", false, 3, writer)
	defer log.Flush()
	log.Info("hello", "world")
}
