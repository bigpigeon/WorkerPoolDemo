package main

import (
	"net/http"
)

var (
	MaxWorker    = 100
	MaxQueue     = 100
	ReqMaxLength = int64(65535)
)

func main() {
	d := AsyncDispatcher(MaxWorker, MaxQueue)
	err := http.ListenAndServe(":8080", http.HandlerFunc(newPayloadHandler(d.Queue, ReqMaxLength)))
	if err != nil {
		panic(err)
	}
}
