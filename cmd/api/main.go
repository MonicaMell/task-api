package main

import (
	//"fmt"
	"log"
	"net/http"
)

func main() {


	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func (w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	})

	port := ":8080"

	log.Printf("server listening on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server failed : %v", err)
	}
}

