package main

import (
	"bufio"
	"fmt"
	"sync"
)

type downloadRequest struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type job struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Status string `json:"status"`
	Output string `json:"output"`
}

type jobStore struct {
	id    int
	store map[int]job
	m     sync.Mutex
}

func (s *jobStore) Put(j job) {
	s.m.Lock()
	s.store[j.ID] = j
	s.m.Unlock()
}

func (s *jobStore) Get(id int) job {
	defer s.m.Unlock()
	s.m.Lock()
	return s.store[id]
}

func (s *jobStore) NextID() int {
	defer s.m.Unlock()
	s.m.Lock()
	s.id++
	return s.id
}

func queue(workers int) (chan<- downloadRequest, *jobStore) {
	q := make(chan downloadRequest)
	s := &jobStore{
		store: map[int]job{},
		m:     sync.Mutex{},
	}

	for i := 0; i < workers; i++ {
		go downloadWorker(q, s)
	}

	return q, s
}

func downloadWorker(q <-chan downloadRequest, s *jobStore) {
	for {
		r := <-q
		j := job{
			ID:     r.ID,
			URL:    r.URL,
			Status: "processing",
		}

		output, err := download(j.URL)
		if err != nil {
			fmt.Printf("can't process job %d: %v\n", j.ID, err)
			j.Status = "failed"
			s.Put(j)
			continue
		}

		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			j.Output += fmt.Sprintf("%s<br>", scanner.Text())
			s.Put(j)
		}
		j.Status = "finished"
		s.Put(j)
	}
}
