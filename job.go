package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

type job struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	Status string `json:"status"`
	Output string `json:"output"`
}

func (j *job) Key() string {
	return JobKey(j.ID)
}

func (j *job) MarshalBinary() (data []byte, err error) {
	return json.Marshal(j)
}

func (j *job) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, j)
}

func (j *job) Save(r *redis.Client) error {
	return r.Set(j.Key(), j, 0).Err()
}

func JobKey(id int64) string {
	return fmt.Sprintf("job:%d", id)
}
