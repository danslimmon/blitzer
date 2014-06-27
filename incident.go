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
}
