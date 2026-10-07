package main

import (
	"net/http"
	"time"
)

func handle_index(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm))
}

func main() {
	http.HandleFunc("/", handle_index)

	http.ListenAndServe(":8888", nil)
}
