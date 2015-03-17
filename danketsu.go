package main

import (
	"flag"
	"strconv"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/status/", statusHandler)
	http.ListenAndServe(":" + strconv.Itoa(*port), nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(status()))
}

func status() string {
	return ""
}

