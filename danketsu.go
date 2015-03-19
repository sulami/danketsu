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

func newEvent(n string) (e *Event) {
	e = new(Event)

	e.Name = n
	e.Timestamp = time.Now()

	return e
}

func fireEvent(n string) {
	cbs := state.callbacks[n]
	if cbs != nil {
		ev := new(Event)
		ev.Name = n
		ev.Timestamp = time.Now()
		state.events = append(state.events, ev)

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

// The global state that is used for testing and statistics.
type State struct {
	// In-memory log of the most recent events.
	events []*Event

	// All registered callbacks.
	callbacks map[string]([]*Callback)
}

var state State = State{
	callbacks: make(map[string]([]*Callback)),
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Maintenance goroutine to clean up the in-memory event log.
	go func() {
		ticker := time.Tick(time.Minute * 5)
		select {
		case <-ticker:
			for i, ev := range(state.events) {
				if time.Since(ev.Timestamp) > time.Hour * 24 {
					state.events = state.events[:i]
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
	return strconv.Itoa(len(state.events))
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
			w.WriteHeader(400)
			return
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
	state.callbacks[n] = append(state.callbacks[n], c)
}

func unregisterCallback(n, a string) {
	cbs := state.callbacks[n]
	if cbs != nil {
		for i, cb := range(cbs) {
			if cb.addr == a {
				// This is some elaborate mechanic to
				// remove an element from a slice
				// without causing a memory leak.
				copy(cbs[i:], cbs[i+1:])
				cbs[len(cbs)-1] = nil
				cbs = cbs[:len(cbs)-1]
				state.callbacks[n] = cbs
			}
		}
	}
}

