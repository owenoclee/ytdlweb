package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gopkg.in/unrolled/render.v1"
)

var b32 = base32.NewEncoding("afhijkoqrtuvwxyzAFHIJKOQRTUVWXYZ")

type downloadRequest struct {
	URL string `json:"url"`
}

type job struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
	Status  string `json:"status"`
	Output  string `json:"output,omitempty"`
}

func newJob(url string) job {
	hash := md5.Sum([]byte(url))
	j := job{
		ID:      b32.EncodeToString(hash[:])[0:8],
		Payload: url,
		Status:  "queued",
	}
	return j
}

type jobStore struct {
	jobs map[string]job
	m    sync.Mutex
}

func newJobStore() *jobStore {
	return &jobStore{
		jobs: map[string]job{},
		m:    sync.Mutex{},
	}
}

func (s *jobStore) put(job job) {
	s.m.Lock()
	s.jobs[job.ID] = job
	s.m.Unlock()
}

func (s *jobStore) get(id string) (job, bool) {
	s.m.Lock()
	job, ok := s.jobs[id]
	s.m.Unlock()

	return job, ok
}

func main() {
	r := chi.NewRouter()
	re := render.New()

	s := newJobStore()
	ctx, cancel := context.WithCancel(context.Background())
	q := queue(ctx, s, 4)

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Post("/download", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		d := json.NewDecoder(r.Body)
		var dlReq downloadRequest
		err := d.Decode(&dlReq)
		if err != nil {
			re.Text(w, http.StatusBadRequest, "bad request")
			return
		}

		job := newJob(dlReq.URL)
		q <- job

		w.Header().Set("Location", fmt.Sprintf("/download/%s", job.ID))
	})

	r.Get("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		job, ok := s.get(chi.URLParam(r, "id"))
		if !ok {
			re.Text(w, http.StatusNotFound, "job not found")
			return
		}

		re.JSON(w, 200, job)
	})

	http.ListenAndServe(":3000", r)
	cancel()
}

func queue(ctx context.Context, s *jobStore, workers int) chan<- job {
	q := make(chan job)

	for i := 0; i < workers; i++ {
		go downloadWorker(ctx, q, s)
	}

	return q
}

func downloadWorker(ctx context.Context, q <-chan job, s *jobStore) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-q:
			fmt.Printf("got job %s\n", job.ID)
			output, err := download(job.Payload)
			if err != nil {
				fmt.Printf("skipping job %s: %v\n", job.ID, err)
				continue
			}
			scanner := bufio.NewScanner(output)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
			fmt.Println("finished scanning")
			job.Status = "processing"
			s.put(job)
		}
	}
}
