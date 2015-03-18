package main

import (
	"flag"
	"strconv"
	"time"
	"encoding/json"
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

var callbacks = map[string]([]*Callback) {
	// Global map to hold all registered callbacks sorted by the
	// callback's name.
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/api/v1/", apiV1Handler)
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

// V1 of the general API - handles everything that will be used by
// other services.
func apiV1Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { // POST for registering a new callback
		type V1Request struct {
			Action, Event, Addr string
		}

		dec := json.NewDecoder(r.Body)
		var rq V1Request
		err := dec.Decode(&rq)
		if err != nil {
			panic(err)
		}

		if rq.Action == "register" {
			registerCallback(rq.Event, r.RemoteAddr)
		}
	}
}

func registerCallback(n string, a string) {
	c := newCallback(n, a)
	callbacks[n] = append(callbacks[n], c)
}

