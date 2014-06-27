package main

func NewSupervisor(pr *ProbeRef) (rsltChan chan *ProbeResult,
                                  ctrlMsgChan chan ProbeSupControlMsg,
                                  ctrlRespChan chan ProbeSupControlResp,
                                  err error) {
    return rsltChan, ctrlMsgChan, ctrlRespChan, nil
}
