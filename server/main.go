package main

import (
	"log"
	"net/http"
)

func main() {
	fileSrv := http.FileServer(http.Dir("./server"))
	if err := http.ListenAndServe(":8080", fileSrv); err != nil {
		log.Fatal(err)
	}
}
