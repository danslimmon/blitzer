package main

import (
    "time"
)

type SupCtrlMsg struct {
    Type string
    Error error
}

type Supervisor struct {
    State string
    ProbeRef *ProbeRef
    RsltChan chan *ProbeResult
    CtrlInChan chan SupCtrlMsg
    CtrlOutChan chan SupCtrlMsg
}

func (sup *Supervisor) Run() {
    Df("Activating probe '%s'", sup.ProbeRef.Name)
    sup.State = "active"
    pd, err := GetProbeDefByName(sup.ProbeRef.Name)
    if err != nil { sup.barf(err) }

    ticker := time.Tick(time.Duration(pd.Interval) * time.Millisecond)
    for sup.State == "active" {
        select {
        case _ = <- ticker:
            sup.kickoffProbe()
        case rslt := <- sup.RsltChan:
            sup.processProbeResult(rslt)
        case msg := <- sup.CtrlInChan:
            sup.processCtrlMsg(msg)
        }   
    }
}

func (sup *Supervisor) Deactivate() {
    sup.CtrlInChan <- SupCtrlMsg{Type:"deactivate"}
}

// Sends an error message out on the channel and kills the Supervisor
func (sup *Supervisor) barf(err error) {
    sup.CtrlOutChan <- SupCtrlMsg{
        Type: "error",
        Error: err,
    }
    sup.Deactivate()
}

func (sup *Supervisor) processCtrlMsg(msg SupCtrlMsg) error {
    switch msg.Type {
    case "deactivate":
        Df("Deactivating probe '%s'", sup.ProbeRef.Name)
        sup.Deactivate()
    }
    return nil
}

func (sup *Supervisor) processProbeResult(rslt *ProbeResult) error {
    D(rslt.Output)
    return nil
}

func (sup *Supervisor) kickoffProbe() error {
    return KickoffProbe(sup.ProbeRef, sup.RsltChan)
}


func NewSupervisor(pr *ProbeRef) (*Supervisor, error) {
    sup := &Supervisor{
        ProbeRef: pr,
        RsltChan: make(chan *ProbeResult),
        CtrlInChan: make(chan SupCtrlMsg),
        CtrlOutChan: make(chan SupCtrlMsg),
    }
    go sup.Run()
    return sup, nil
}
