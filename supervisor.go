package main

type SupControlMsg struct {
    Type string
    Error error
}

type Supervisor struct {
    State string
    ProbeRef *ProbeRef
    RsltChan chan *ProbeResult
    CtrlInChan chan SupControlMsg
    CtrlOutChan chan SupControlMsg
}

func (sup *Supervisor) Run() {
    sup.State = "active"
    _, err := GetProbeDefByName(sup.ProbeRef.Name)
    if err != nil { sup.barf(err) }
    _ = <- sup.CtrlInChan
}

func (sup *Supervisor) Deactivate() {
    sup.CtrlInChan <- SupControlMsg{Type:"deactivate"}
}

// Sends an error message out on the channel and kills the Supervisor
func (sup *Supervisor) barf(err error) {
    sup.CtrlOutChan <- SupControlMsg{
        Type: "error",
        Error: err,
    }
    sup.Deactivate()
}

func NewSupervisor(pr *ProbeRef) (*Supervisor, error) {
    sup := &Supervisor{
        ProbeRef: pr,
        RsltChan: make(chan *ProbeResult),
        CtrlInChan: make(chan SupControlMsg),
        CtrlOutChan: make(chan SupControlMsg),
    }
    go sup.Run()
    return sup, nil
}
