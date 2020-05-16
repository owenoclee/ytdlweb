package main

import (
	"io"
	"os/exec"
)

func download(url string) (io.Reader, error) {
	ytdl := exec.Command(
		"youtube-dl",
		"--newline",
		"--all-subs",
		"-o 'downloads/%(title)s-%(id)s.%(ext)s'",
		url,
	)

	stdout, err := ytdl.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := ytdl.StderrPipe()
	if err != nil {
		return nil, err
	}
	ytdl.Start()

	return io.MultiReader(stdout, stderr), nil
}
