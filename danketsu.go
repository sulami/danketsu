package main

import (
	"flag"
	"strconv"
	"time"
	"net/http"
)

type Event struct {
	name      string // "prefix_event"
	timestamp time.Time
}

func newEvent(n string) (e *Event) {
	e = new(Event)

	e.name = n
	e.timestamp = time.Now()

	return e
}

type Callback struct {
	event string // "prefix_event"
	addr  string // "http://1.2.3.4:56/ev/"
}

func newCallback(e string, a string) (c *Callback) {
	c = new(Callback)

	c.event = e
	c.addr = a

	return c
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/status/", statusHandler)
	http.ListenAndServe(":" + strconv.Itoa(*port), nil)
}

// We use handlers that just write the return of the corrsoponding
// functions, which makes testing a lot simpler.

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(status()))
}

func status() string {
	return ""
}

