### 引言


这个demo是根据**使用Go语言每分钟处理1百万请求**[译文](https://github.com/itfanr/articles-about-golang/blob/master/2016-10/1.handling-1-million-requests-per-minute-with-golang.md)[原文](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)
中的代码写的，目的是为了测试线程池的瓶颈和缺陷


而且我发现代码中以下部分会导致创建过多的gorouter,所以把它从go func中提到外面去,通过增大queue的方法来缓存任务

```
go func(job Job) {
    // try to obtain a worker job channel that is available.
    // this will block until a worker is idle
        jobChannel := <-d.WorkerPool

    // dispatch the job to the worker job channel
    jobChannel <- job
}(job)
```


### 如何使用

    go get -d github.com/bigpigeon/WorkerPoolDemo
	
	cd $GOPATH/src/github.com/bigpigeon/WorkerPoolDemo

    go test -v -race
	

然后你将看到一下信息
```
2017/01/24 12:00:24 19 th request 3 th task finish in 504.4885ms
2017/01/24 12:00:24 22 th request 2 th task finish in 508.440571ms
2017/01/24 12:00:24 38 th request 3 th task finish in 516.64915ms
...

--- PASS: TestRequest (2.12s)
	request_test.go:99: MaxWorker:200 MaxQueue:200 ReqInterval:0 ReqTimes:50 ReqPreloadLen:5 ReqPreloadLenFloat:0 ReqPreloadWait:500ms ReqPreloadWaitFloat:1s GOMAXPROCS:4 
=== RUN   TestBlock
--- PASS: TestBlock (13.14s)
	request_test.go:120: MaxWorker:50 MaxQueue:50 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:10 ReqPreloadLenFloat:0 ReqPreloadWait:100ms ReqPreloadWaitFloat:0 GOMAXPROCS:4 
	request_test.go:123: not block
	request_test.go:120: MaxWorker:50 MaxQueue:50 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:10 ReqPreloadLenFloat:0 ReqPreloadWait:150ms ReqPreloadWaitFloat:0 GOMAXPROCS:4 
	request_test.go:123: not block
	request_test.go:120: MaxWorker:50 MaxQueue:50 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:10 ReqPreloadLenFloat:0 ReqPreloadWait:200ms ReqPreloadWaitFloat:0 GOMAXPROCS:4 
	request_test.go:123: not block
	request_test.go:120: MaxWorker:50 MaxQueue:50 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:10 ReqPreloadLenFloat:0 ReqPreloadWait:250ms ReqPreloadWaitFloat:0 GOMAXPROCS:4 
	request_test.go:123: not block
	request_test.go:120: MaxWorker:50 MaxQueue:50 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:10 ReqPreloadLenFloat:0 ReqPreloadWait:300ms ReqPreloadWaitFloat:0 GOMAXPROCS:4 
	request_test.go:125: block at 35 th request
=== RUN   TestProfiling
--- PASS: TestProfiling (6.28s)
	request_test.go:146: MaxWorker:500 MaxQueue:500 ReqInterval:50ms ReqTimes:50 ReqPreloadLen:50 ReqPreloadLenFloat:0 ReqPreloadWait:500ms ReqPreloadWaitFloat:1s GOMAXPROCS:4 
	request_test.go:158: not block
PASS
ok  	_/D_/work/GoDemo/src/WorkerPoolDemo	21.608s
```

**参数说明**

```
MaxWorker 工作线程数
MaxQueue  最大队列数
ReqInterval 请求间隔
ReqTimes    请求总数
ReqPreloadLen 每次请求包含的任务数
ReqPreloadLenFloat 任务数的浮动值(总任务数=ReqPreloadLen + random(0-ReqPreloadLenFloat))
ReqPreloadWait 任务所需的处理时间
ReqPreloadWaitFloat 任务处理时间的浮动值(总时间=ReqPreloadWait + random(0-ReqPreloadWaitFloat))
GOMAXPROCS 真实线程数(这个线程不是routing而是runtime.GOMAXPROCS，默认等于cpu的数量)
```


**查看Cpu profile**

    go test -cpuprofile=demo.prof
	go tool pprof WorkerPoolDemo.test demo.prof
	
	
在windows中执行文件是WorkerPoolDemo.test.exe
	
**查看trace**

	go test -trace=demo.trace
	go tool trace WorkerPoolDemo.test demo.trace

在windows中执行文件是WorkerPoolDemo.test.exe


如果发现trace的网页无法正常显示请更新go或者浏览器版本
