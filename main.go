package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"syscall"
)

func main() {
	http.HandleFunc("/", Handler)
	syscall.Mkfifo("./fifo", 0o750)
	log.Fatal(http.ListenAndServe(":9924", nil))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New request", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	fifo, err := os.OpenFile("./fifo", os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(200)
	tee := io.TeeReader(fifo, w)
	buf := make([]byte, 8 * 1024 * 1024)
	for {
		_, err := tee.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	fifo.Close()
}
