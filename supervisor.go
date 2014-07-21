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
    RsltInChan chan *ProbeResult
    RsltOutChan chan *ProbeResult
    CtrlInChan chan SupCtrlMsg
    CtrlOutChan chan SupCtrlMsg
}

func (sup *Supervisor) Run() {
    Df("Activating probe '%s'", sup.ProbeRef.Name)
    sup.State = "active"
    pd, err := GetProbeDefByName(sup.ProbeRef.Name)
    if err != nil { sup.barf(err) }

    ticker := time.Tick(time.Duration(pd.Interval) * time.Second)
    for sup.State == "active" {
        select {
        case _ = <- ticker:
            sup.kickoffProbe()
        case rslt := <- sup.RsltInChan:
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
    Df("Got result from probe '%s': %s", rslt.Ref.Name, rslt.Values)
    sup.RsltOutChan <- rslt
    return nil
}

func (sup *Supervisor) kickoffProbe() error {
    return KickoffProbe(sup.ProbeRef, sup.RsltInChan)
}


func NewSupervisor(pr *ProbeRef, rsltOutChan chan *ProbeResult) (*Supervisor, error) {
    sup := &Supervisor{
        ProbeRef: pr,
        RsltInChan: make(chan *ProbeResult),
        RsltOutChan: rsltOutChan,
        CtrlInChan: make(chan SupCtrlMsg),
        CtrlOutChan: make(chan SupCtrlMsg),
    }
    go sup.Run()
    return sup, nil
}
