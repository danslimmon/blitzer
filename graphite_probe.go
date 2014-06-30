package main

type GraphiteConf struct {
    BaseURL string
}

type GraphiteProbe struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult
}

func (p *GraphiteProbe) Init(ref *ProbeRef, def *ProbeDef, ch chan *ProbeResult) {
    p.Ref = ref
    p.Def = def
    p.RsltChan = ch
}

func (p *GraphiteProbe) Kickoff() error {
    Df("Kicking off Graphite probe '%s' from file '%s'", p.Def.Name, p.Ref.SourceFile)
    go p.kickoff()
    return nil
}

func (p *GraphiteProbe) kickoff() {
    // Execute template here
    p.RsltChan <- &ProbeResult{}
}
