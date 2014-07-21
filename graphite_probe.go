package main

import (
    "fmt"
    "time"
    "strings"
    "text/template"
)

type GraphiteConf struct {
    BaseURL string
}

type GraphiteProbe struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult
}

type stringWriter struct {S string}
func (sw *stringWriter) Write(p []byte) (n int, err error) {
    sw.S = strings.Join([]string{sw.S, string(p)}, "")
    return len(p), nil
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
    rslt := new(ProbeResult)
    rslt.Ref = p.Ref
    rslt.Success = true
    rslt.Values = make(map[string]string, 0)
    rslt.Timestamp = time.Now().Unix()

    url, err := p.makeGraphiteURL()
    if err != nil {
        rslt.Success = false
        rslt.Error = err
        p.RsltChan <- rslt
        return
    }
    rslt.Values["img_url"] = url

    p.RsltChan <- rslt
}

func (p *GraphiteProbe) makeGraphiteURL() (string, error) {
    parts := make([]string, 2)
    parts[0] = Config.Graphite.BaseURL
    qs, err := p.makeQueryString()
    if err != nil { return "", err }
    parts[1] = qs
    return strings.Join(parts, "?"), nil
}

func (p *GraphiteProbe) makeQueryString() (string, error) {
    qs_template, ok := p.Def.Args["qs_template"].(string)
    if ! ok {
        return "", ConfigurationError{fmt.Sprintf("Graphite probe definition in '%s' requires 'qs_template' parameter which must be a string", p.Def.SourceFile)}
    }
    tmpl, err := template.New("graphite_qs").Parse(qs_template)
    writer := &stringWriter{}
    ctx := ProbeContext(p.Def, p.Ref)
    err = tmpl.Execute(writer, ctx)
    return writer.S, err
}
