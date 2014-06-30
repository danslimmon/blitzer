package main

import (
    "fmt"
)

type ProbeResult struct {
    Ref *ProbeRef
    Success bool
    Output string
    Error error
}

type ProbeRun interface {
    Kickoff() error
}

func KickoffProbe(probeRef *ProbeRef, rsltChan chan *ProbeResult) error {
    probeDef, err := GetProbeDefByName(probeRef.Name)
    if err != nil { return err }
    var run ProbeRun
    switch probeDef.Type {
    case "ansible":
        run = &AnsibleProbeRun{
            Ref: probeRef,
            Def: probeDef,
            RsltChan: rsltChan,
        }
    default:
        return ConfigurationError{fmt.Sprintf("Unknown probe type '%s'", probeDef.Type)}
    }
    go run.Kickoff()
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
