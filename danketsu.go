package main

import (
	"flag"
	"strconv"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.ListenAndServe(":" + strconv.Itoa(*port), nil)
}

