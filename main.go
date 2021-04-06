package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"syscall"
)

var port = "9924"
var fifoPath = "./fifo"

const bufSize = 8 * 1024 * 1024

func main() {
	if f, exists := os.LookupEnv("FIFO_PATH"); exists {
		fifoPath = f
	}
	if p, exists := os.LookupEnv("PORT"); exists {
		port = p
	}

	http.HandleFunc("/", ReceiveHandler)
	http.HandleFunc("/receive", ReceiveHandler)
	http.HandleFunc("/transmit", TransmitHandler)
	if err := syscall.Mkfifo(fifoPath, 0o750); err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

var mut sync.Mutex

func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Waiting for lock...")
	mut.Lock()
	defer mut.Unlock()
	fmt.Println("New request (receive):", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	fifo, err := os.OpenFile("./fifo", os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(200)
	if _, err := io.CopyBuffer(w, fifo, make([]byte, bufSize)); err != nil {
		fmt.Println(err)
	}
	_ = fifo.Close()
}

func TransmitHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Waiting for lock...")
	mut.Lock()
	defer mut.Unlock()
	fmt.Println("New request (transmit):", r.RemoteAddr)
	fifo, err := os.OpenFile("./fifo", os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(200)
	if _, err := io.CopyBuffer(fifo, r.Body, make([]byte, bufSize)); err != nil {
		fmt.Println(err)
	}
	_ = fifo.Close()
}
