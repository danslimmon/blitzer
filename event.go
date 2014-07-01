package blitzer

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
        return &Event{}, WebError{Code: 500, Message: "Error reading request body"}
    }

    ne := new(nagiosEvent)
    err = json.Unmarshal(body, ne)
    if err != nil {
        return &Event{}, WebError{Code: 400, E: err}
    }

    event := new(Event)
    if ne.ServiceName == "" {
        e := WebError{Code: 400, Message: "Missing 'service' parameter"}
        log.Println("wat")
        return &Event{}, e
    } else {
        event.ServiceName = ne.ServiceName
    }

    event.ServiceName = ne.ServiceName
    switch ne.State {
    case "CRITICAL":
        event.State = "down"
    case "OK":
        event.State = "up"
    case "":
        return &Event{}, nil
    }

    return event, nil
}
