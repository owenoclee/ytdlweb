package main

//go:generate go run pack_public.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"gopkg.in/unrolled/render.v1"
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	workerCount, _ := strconv.ParseInt(os.Getenv("WORKER_COUNT"), 10, 32)
	if listenAddr == "" {
		listenAddr = ":3000"
	}
	if workerCount < 1 {
		workerCount = 4
	}

	queue, jobStore := queue(int(workerCount))

	mux := chi.NewRouter()
	ren := render.New()

	mux.Get("/", http.FileServer(publicAssets).ServeHTTP)

	mux.Post("/download", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		var dlReq downloadRequest
		err := json.NewDecoder(r.Body).Decode(&dlReq)
		if err != nil {
			ren.Text(w, http.StatusBadRequest, "unable to decode input")
			return
		}

		dlReq.ID = jobStore.NextID()
		queue <- dlReq

		w.Header().Set("Location", fmt.Sprintf("/download/%d", dlReq.ID))
	})

	mux.Get("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
		if err != nil {
			ren.Text(w, http.StatusBadRequest, "given id is invalid")
			return
		}

		j := jobStore.Get(int(id))
		if j.ID == 0 {
			ren.Text(w, http.StatusNotFound, "not found")
			return
		}

		ren.JSON(w, 200, j)
	})

	panic(http.ListenAndServe(listenAddr, mux))
}
