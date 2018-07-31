# elog high perfermance logger lib for golang
<pre>
func main() {
	flag.Parse()
	defer elog.Flush()
	elog.Info("hello","world")
  	elog.Infof("hello %s","world")  
}
</pre>
./main -logToStderr -logLevel=INFO -logFlushTime=3 -logPath=./
