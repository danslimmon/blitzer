package main

import (
    "fmt"
)

type ProbeResult struct {
    Ref *ProbeRef
    Success bool
    Values map[string]string
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
    case "ansible":
        p = &AnsibleProbe{}
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
