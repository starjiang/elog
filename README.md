# elog high perfermance logger lib for golang
<pre>
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
</pre>
./example_elog -logToStderr -logLevel=INFO -logFlushTime=3 -logPath=./

Handler
=======
<pre>
type TestLogHandler struct {
}

func (lh *TestLogHandler) Write(data []byte) (int, error) {

	return os.Stderr.Write(data)
}

func (lh *TestLogHandler) Flush() {

}

func main() {
	writer := &TestLogHandler{}
	log := elog.NewEasyLogger("INFO", false, 3, writer)
	log.Info("hello", "world")
}
</pre>
