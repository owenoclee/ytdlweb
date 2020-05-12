package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"gopkg.in/unrolled/render.v1"
)

func main() {
	mux := chi.NewRouter()
	ren := render.New()
	rds := redis.NewClient(&redis.Options{})

	go queue(rds, 4)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

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
		}

		ren.JSON(w, 200, job)
	})

	http.ListenAndServe(":3000", mux)
}
