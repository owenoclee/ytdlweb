package main

//go:generate go run pack_public.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"gopkg.in/unrolled/render.v1"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	listenAddr := os.Getenv("LISTEN_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if listenAddr == "" {
		listenAddr = ":80"
	}

	mux := chi.NewRouter()
	ren := render.New()
	rds := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := rds.Ping().Result(); err != nil {
		panic(err)
	}

	go queue(rds, 4)

	mux.Get("/", http.FileServer(publicAssets).ServeHTTP)

	mux.Post("/download", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		job := &job{}
		err := json.NewDecoder(r.Body).Decode(job)
		if err != nil {
			ren.Text(w, http.StatusBadRequest, "unable to decode input")
			return
		}

		id, err := rds.Incr("job_id").Result()
		if err != nil {
			ren.Text(w, http.StatusInternalServerError, "failed on redis")
			return
		}

		job.ID = id
		job.Status = "queued"
		job.Output = ""

		rds.Set(job.Key(), job, 0)
		rds.Publish("download_queue", job.Key())

		w.Header().Set("Location", fmt.Sprintf("/download/%d", job.ID))
	})

	mux.Get("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			ren.Text(w, http.StatusBadRequest, "given id is invalid")
			return
		}

		job := &job{}
		err = rds.Get(JobKey(id)).Scan(job)
		if err != nil {
			ren.Text(w, http.StatusNotFound, "job not found")
			return
		}

		ren.JSON(w, 200, job)
	})

	http.ListenAndServe(listenAddr, mux)
}
