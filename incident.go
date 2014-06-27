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

    // Communication channels with the incident's running ProbeSups.
    //
    // Each key is derived from ProbeRef.Hash(). The channels are defined
    // as follows:
    //
    //  SupRsltChans: The ProbeSupervisor will send any new
    //      ProbeResults on this channel.
    //  SupCtrlMsgChans: This channel is used to send control messages
    //      to a ProbeSupervisor, e.g. to activate or deactive it.
    //  SupCtrlRespChans: This channel carries the responses to control
    //      messages that were sent on SupCtrlMsgChans channels.
    SupRsltChans map[string]chan *ProbeResult
    SupCtrlMsgChans map[string]chan ProbeSupControlMsg
    SupCtrlRespChans map[string]chan ProbeSupControlResp
}

func NewIncident(event *Event, probeRefs []*ProbeRef) (*Incident, error) {
    inc := new(Incident)
    inc.State = "active"
    inc.ProbeRefs = probeRefs
    inc.History = make([]*HistoryEvent, 0)

    for _, pr := range inc.ProbeRefs {
        Df("Creating new supervisor for incident '%s'\n", event.ServiceName)
        rsltChan, ctrlMsgChan, ctrlRespChan, err := NewSupervisor(pr)
        if err != nil { return &Incident{}, nil }
        inc.SupRsltChans[pr.Hash()] = rsltChan
        inc.SupCtrlMsgChans[pr.Hash()] = ctrlMsgChan
        inc.SupCtrlRespChans[pr.Hash()] = ctrlRespChan
    }

    return inc, nil
}
