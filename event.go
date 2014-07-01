package main

import (
    "log"
    "net/http"
    "encoding/json"
    "io/ioutil"
)

type Event struct {
    ServiceName string
    State string
}

// The data structure into which we deserialize a Nagios event
type nagiosEvent struct {
    ServiceName string `json:"service"`
    State string
}

func NewEventFromNagios(r *http.Request) (*Event, error) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Fatal(err)
        return &Event{}, WebError{Code: 500, Message: "Error reading request body"}
    }

    ne := new(nagiosEvent)
    err = json.Unmarshal(body, ne)
    if err != nil { return &Event{}, nil }

    event := new(Event)
    event.ServiceName = ne.ServiceName
    switch ne.State {
    case "CRITICAL":
        event.State = "down"
    case "OK":
        event.State = "up"
    }

    return event, nil
}
