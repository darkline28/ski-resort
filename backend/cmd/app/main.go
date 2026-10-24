package main

import (
	"fmt"
	"log"
	"net/http"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Ski Resort!")
}

func main() {
	http.HandleFunc("GET /hello", Hello)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
