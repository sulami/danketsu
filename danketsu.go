package main

import (
	"bytes"
	"flag"
	"strconv"
	"time"
	"encoding/json"
	"net/http"
)

type Event struct {
	Name      string // "prefix_event"
	Timestamp time.Time
}

var events = []*Event {
	// This is a list of event pointers that is used as in-memory
	// short-time log and of course for statistics.
}

func newEvent(n string) (e *Event) {
	e = new(Event)

	e.Name = n
	e.Timestamp = time.Now()

	return e
}

func fireEvent(n string) {
	cbs := callbacks[n]
	if cbs != nil {
		ev := new(Event)
		ev.Name = n
		ev.Timestamp = time.Now()
		events = append(events, ev)

		for _, cb := range cbs {
			fire(ev, cb.addr)
		}
	}
}

func fire(e *Event, addr string) {
	enc, err := json.Marshal(e)
	if err != nil {
		// TODO log failure
		return
	}
	http.Post(addr, "application/json", bytes.NewBuffer(enc))
}

type Callback struct {
	event string // "prefix_event"
	addr  string // "http://1.2.3.4:56/ev/"
}

func newCallback(e, a string) (c *Callback) {
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

	// Maintenance goroutine to clean up the in-memory event log.
	go func() {
		ticker := time.Tick(time.Minute * 5)
		select {
		case <-ticker:
			for i, ev := range(events) {
				if time.Since(ev.Timestamp) > time.Hour * 24 {
					events = events[:i]
					break
				}
			}
		}
	} ()

	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/api/v1/", apiV1Handler)
	http.ListenAndServe(":" + strconv.Itoa(*port), nil)
}

// We use handlers that just write the return of the corrsoponding
// functions, which makes testing a lot simpler.

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(status()))
}

// Return a set of statistics about the service for monitoring reasons.
func status() string {
	return ""
}

// V1 of the general API - handles everything that will be used by
// other services.
func apiV1Handler(w http.ResponseWriter, r *http.Request) {
	// POST for (un)registering a callback or firing events.
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
		} else if rq.Action == "fire" {
			fireEvent(rq.Event)
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

