package main

import (
	"bufio"
	"fmt"

	"github.com/go-redis/redis"
)

func queue(r *redis.Client, workers int) {
	todo := make(chan *job)

	for i := 0; i < workers; i++ {
		go downloadWorker(r, todo)
	}

	pubsub := r.Subscribe("download_request")
	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}

	queue := pubsub.Channel()
	for {
		nextKey := <-queue

		nextJob := &job{}
		err := r.Get(nextKey.Payload).Scan(nextJob)
		if err != nil {
			fmt.Printf("can't decode job %s: %v\n", nextKey.Payload, err)
			return
		}

		todo <- nextJob
	}
}

func downloadWorker(r *redis.Client, todo <-chan *job) {
	for {
		job := <-todo

		output, err := download(job.URL)
		job.Status = "processing"
		if err != nil {
			fmt.Printf("can't process job %d: %v\n", job.ID, err)
			job.Status = "failed"
			job.Save(r)
			continue
		}

		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			job.Output += fmt.Sprintf("%s<br>", scanner.Text())
			job.Save(r)
		}
		job.Status = "finished"
		job.Save(r)
	}
}
