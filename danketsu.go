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
	// POST for (un)registering a callback
	if r.Method == "POST" {
		type V1Request struct {
			Action, Event, Address string
		}

		dec := json.NewDecoder(r.Body)
		var rq V1Request
		err := dec.Decode(&rq)
		if err != nil {
			panic(err)
		}

		if rq.Action == "register" {
			registerCallback(rq.Event, rq.Address)
		} else if rq.Action == "unregister" {
			unregisterCallback(rq.Event, rq.Address)
		} else {
			w.WriteHeader(400)
		}
	}
}

func registerCallback(n, a string) {
	c := newCallback(n, a)
	callbacks[n] = append(callbacks[n], c)
}

func unregisterCallback(n, a string) {
	cbs := callbacks[n]
	if cbs != nil {
		for i, cb := range(cbs) {
			if cb.addr == a {
				// This is some elaborate mechanic to
				// remove an element from a slice
				// without causing a memory leak.
				copy(cbs[i:], cbs[i+1:])
				cbs[len(cbs)-1] = nil
				cbs = cbs[:len(cbs)-1]
				callbacks[n] = cbs
			}
		}
	}
}

