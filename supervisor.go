package main

type Supervisor struct {
    ProbeRef *ProbeRef
    RsltChan chan *ProbeResult
    CtrlMsgChan chan ProbeSupControlMsg
    CtrlRespChan chan ProbeSupControlResp
}

func NewSupervisor(pr *ProbeRef) (*Supervisor, error) {
    sup := &Supervisor{
        ProbeRef: pr,
        RsltChan: make(chan *ProbeResult),
        CtrlMsgChan: make(chan ProbeSupControlMsg),
        CtrlRespChan: make(chan ProbeSupControlResp),
    }
    return sup, nil
}
