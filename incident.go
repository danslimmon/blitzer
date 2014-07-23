package main

import (
    "github.com/fzzy/radix/redis"
)

type HistoryEvent struct {
    Timestamp int64 `json:"timestamp"`
    ProbeName string `json:"probe_name"`
    Success bool `json:"success"`
    Values map[string]string `json:"values"`
}

type IncCtrlMsg struct {
    Type string
    Error error
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
    Supervisors map[string]*Supervisor
    RsltChan chan *ProbeResult
    Slug string
    RedisClient *redis.Client

    ctrlChan chan *IncCtrlMsg
}

func (inc *Incident) Run() {
    for {
        select {
        case ctrlMsg := <- inc.ctrlChan:
            if ctrlMsg.Type == "deactivate" {
                Df("Stopping incident '%s'", inc.Slug)
                return
            }
        case pr := <- inc.RsltChan:
            Df("Saving probe result for '%s' to database", pr.Ref.Name)
            err := inc.writeProbeResult(pr)
            if err != nil { Df("Unable to write probe result for '%s' to database: %s", pr.Ref.Name, err) }
        }
    }
}

func (inc *Incident) writeProbeResult(pr *ProbeResult) error {
    he := HistoryEvent{
        Timestamp: pr.Timestamp,
        ProbeName: pr.Ref.Name,
        Success: pr.Success,
        Values: pr.Values,
    }

    db, err := GetDB()
    if err != nil { return err }
    err = db.WriteHistory(inc, he)
    if err != nil { return err }
    return nil
}

func (inc *Incident) Deactivate() {
    for _, sup := range inc.Supervisors {
        sup.Deactivate()
    }
    inc.ctrlChan <- &IncCtrlMsg{Type: "deactivate"}
    inc.State = "inactive"
}

func NewIncident(event *Event, triggerDefs []*TriggerDef) (*Incident, error) {
    inc := new(Incident)
    inc.State = "active"
    inc.ProbeRefs = make([]*ProbeRef, 0)
    inc.Supervisors = make(map[string]*Supervisor)
    inc.RsltChan = make(chan *ProbeResult)
    inc.Slug = "2014-07-21_fake_incident"

    for _, td := range triggerDefs {
        for _, pr := range td.ProbeRefs {
            inc.ProbeRefs = append(inc.ProbeRefs, pr)
        }
    }

    for _, pr := range inc.ProbeRefs {
        Df("Creating new supervisor for incident '%s'\n", event.ServiceName)
        sup, err := NewSupervisor(pr, inc.RsltChan)
        if err != nil { return &Incident{}, nil }
        inc.Supervisors[pr.Hash()] = sup
    }

    go inc.Run()
    return inc, nil
}
