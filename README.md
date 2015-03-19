団結
==

*Danketsu*

A simple event-based microservice communication microservice. Operates on an
user-defined port over HTTP using JSON. Also insanely fast, typical requests
take about 50ms on a local workstation.

Usage
-----

Compile with `go build`.

The binary takes an optional parameter `-port`, which defaults to 8080.

The APIv1 can be reached via HTTP POST at `http://server:port/api/v1/` and
defines the following interactions:

- Registering a callback

    {
        "action": "register",
        "event": "users_new_user_created",
        "address": "http://deimos.company.local:1338/api/v1/"
    }

- Unregistering a callback

    {
        "action": "unregister",
        "event": "users_new_user_created",
        "address": "http://deimos.company.local:1338/api/v1/"
    }

- Firing an event

    {
        "action": "fire",
        "event": "users_new_user_created"
    }

The addresses passed are the ones for the callbacks, so Danketsu knows who to
message in case an event occurs.

There is also a status page at `http://server:port/status/` which currently
returns the number of fired events in the last 24 hours when called via HTTP
GET.

There is no internal event database, which means events are not more than
arbitrary strings. You can register for events that do no exists and can fire
events that are not registered anywhere.

The server will return HTTP status code 200 in any case, except when the input
data is malformed, in which case it will return HTTP status code 400.

