# elog high perfermance logger lib for golang
global logger model
==================
```
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
```
./example_elog -logToStderr -logLevel=INFO -logFlushTime=3 -logPath=./

LoggerHandler model
===================
```
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
```
elog config
=================
```
-logFlushTime int
    	log flush time interval,default 3 seconds (default 3)
  -logLevel string
    	log level[DEBUG,INFO,WARN,ERROR,NONE],default INFO level (default "INFO")
  -logPath string
    	log path,default log to current directory (default "./")
  -logToStderr
    	log to stderr,default false
```
elog rotate file rules
======================
```
1.when file size reach to 1GB,the logger file will be rotate
2.rotate files max num is 10
```
elog high performance
======================
```
log to cache buffer then flush to file
cache buffer size is 1M
default flush time 3 seconds,-logFlushTime can change flush time interval
if want close log outputing,-logLevel=NONE can close log outputing
```

