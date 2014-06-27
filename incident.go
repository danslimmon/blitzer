package main

import (
    "time"
)

type HistoryEvent struct {
    Time time.Time
    ProbeResult *ProbeResult
}

// An Incident represents a set of probes and their result history.
//
// When an Incident is activated by an Event, its probes start running. When
// it is deactivated by another Event, its probes stop running. The output
// of all probes is stored in the Incident.
type Incident struct {
    // "active" or "inactive"
    State string
    ProbeRefs []*ProbeRef
    History []*HistoryEvent
    Supervisors map[string]*Supervisor
}

func NewIncident(event *Event, probeRefs []*ProbeRef) (*Incident, error) {
    inc := new(Incident)
    inc.State = "active"
    inc.ProbeRefs = probeRefs
    inc.History = make([]*HistoryEvent, 0)

    for _, pr := range inc.ProbeRefs {
        Df("Creating new supervisor for incident '%s'\n", event.ServiceName)
        sup, err := NewSupervisor(pr)
        if err != nil { return &Incident{}, nil }
        inc.Supervisors[pr.Hash()] = sup
    }

    return inc, nil
}
