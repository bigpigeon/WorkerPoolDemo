package main

import (
	"log"
	"time"
)

type Dispatcher struct {
	WorkerPool    chan chan Job
	Queue         chan Job
	MaxWorkerPool int
	MaxQueue      int
}

func NewDispatcher(maxWorkers int, maxQueue int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	queue := make(chan Job, maxQueue)
	return &Dispatcher{
		WorkerPool:    pool,
		Queue:         queue,
		MaxWorkerPool: maxWorkers,
		MaxQueue:      maxQueue,
	}
}

func AsyncDispatcher(max_workers int, max_queue int) *Dispatcher {
	d := NewDispatcher(max_workers, max_queue)
	d.Run()
	go d.dispatch()
	// await until all worker ready
	d.Join()
	return d
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.MaxWorkerPool; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}
	log.Println("Dispatcher await...")
}

func (d *Dispatcher) Stop() {
	d.Join()
	close(d.Queue)
	for len(d.WorkerPool) > 0 {
		jobChannel := <-d.WorkerPool
		close(jobChannel)
	}
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job, more := <-d.Queue:
			if more == false {
				return
			}
			jobChannel := <-d.WorkerPool
			jobChannel <- job
		}
	}
}

func (d *Dispatcher) IsBlock() bool {
	//	log.Printf("free workers:%d, queue length: %d", len(d.WorkerPool), len(d.Queue))
	return len(d.WorkerPool) == 0 && len(d.Queue) == (d.MaxQueue-1)
}

func (d *Dispatcher) Join() {
	for {
		if len(d.WorkerPool) == d.MaxWorkerPool && len(d.Queue) == 0 {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}
