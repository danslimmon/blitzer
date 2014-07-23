package main

import (
    "fmt"
)

type ProbeResult struct {
    Def *ProbeDef
    Ref *ProbeRef
    Success bool
    Values map[string]string
    Timestamp int64
    Error error
}

type Probe interface {
    Init(*ProbeRef, *ProbeDef, chan *ProbeResult)
    Kickoff() error
}

func KickoffProbe(probeRef *ProbeRef, rsltChan chan *ProbeResult) error {
    probeDef, err := GetProbeDefByName(probeRef.Name)
    if err != nil { return err }
    var p Probe
    switch probeDef.Type {
    case "shell":
        p = &ShellProbe{}
    case "graphite":
        p = &GraphiteProbe{}
    default:
        return ConfigurationError{fmt.Sprintf("Unknown probe type '%s'", probeDef.Type)}
    }
    p.Init(probeRef, probeDef, rsltChan)
    go p.Kickoff()
    return nil
}

func GetProbeDefByName(name string) (*ProbeDef, error) {
    for _, pd := range Config.ProbeDefs {
        if pd.Name == name {
            return pd, nil
        }
    }
    return nil, ConfigurationError{fmt.Sprintf("Cannot find probe with name '%s'", name)}
}

func ProbeContext(def *ProbeDef, ref *ProbeRef) map[string]interface{} {
    ctx := make(map[string]interface{}, 0)
    for k, v := range def.Args { ctx[k] = v }
    for k, v := range ref.Args { ctx[k] = v }
    return ctx
}
