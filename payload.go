package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type Payload struct {
	Wait     time.Duration `json:"wait"`
	ReqIndex int64
	Index    int
}

func (p *Payload) Run() error {
	time.Sleep(p.Wait)
	log.Printf("%d th request %d th task finish in %v", p.ReqIndex, p.Index, p.Wait)
	return nil
}

type PayloadCollection []Payload
type Job struct {
	Payload Payload
}

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) *Worker {
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			//register the current worker
			w.WorkerPool <- w.JobChannel
			select {
			case job, more := <-w.JobChannel:
				if more == false {
					//					log.Println("worker stop")
					return
				}
				//we have received a work request
				if err := job.Payload.Run(); err != nil {
					log.Printf("Error task :%s\n", err.Error())
				}
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func newPayloadHandler(queue chan Job, max_length int64) func(w http.ResponseWriter, r *http.Request) {
	var index int64
	return func(w http.ResponseWriter, r *http.Request) {
		curr_index := atomic.AddInt64(&index, 1)
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var content PayloadCollection
		err := json.NewDecoder(io.LimitReader(r.Body, max_length)).Decode(&content)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err.Error())
			return
		}
		for i, payload := range content {
			payload.ReqIndex = curr_index - 1
			payload.Index = i
			work := Job{Payload: payload}
			queue <- work
		}
		w.WriteHeader(http.StatusOK)
	}
}
