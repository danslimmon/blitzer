package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "io/ioutil"
    "launchpad.net/goyaml"
)

type ConfigurationError struct { s string }
func (e ConfigurationError) Error() string { return e.s }

type AnsibleConf struct {
    Inventory string
}

type BlitzerConf struct {
    Addr string
    Port int16
    Debug string
    Ansible AnsibleConf
    TriggerDefs []*TriggerDef
    ProbeDefs []*ProbeDef
}

type ProbeRef struct {
    Name string
    Args map[string]string
}

type TriggerDef struct {
    ServiceMatch string "service_match"
    Probes []*ProbeRef
}

type AnsibleProbeDef struct {
    Module string
    Args string
}

type ProbeDef struct {
    Name string
    Type string
    Html string

    // Types of probe we support
    Ansible AnsibleProbeDef
}

// Loads our configuration from the given path
func GetConf(yamlPath string) (BlitzerConf, error) {
    f, err := os.Open(yamlPath)
    if err != nil { return BlitzerConf{}, err }
    defer f.Close()

    yamlBytes, err := ioutil.ReadAll(f)
    if err != nil { return BlitzerConf{}, err }

    conf := new(BlitzerConf)
    goyaml.Unmarshal(yamlBytes, conf)
    if err != nil { return BlitzerConf{}, err }
    if conf.Debug == "yes" {
        log.Printf("Loaded main configuration from '%s'\n", yamlPath)
    }

    // Load the triggers.d/*.yaml files that describe individual triggers.
    triggerPaths, err := filepath.Glob(filepath.Join(filepath.Dir(yamlPath), "triggers.d", "*.yaml"))
    if err != nil { return BlitzerConf{}, err }
    if len(triggerPaths) == 0 {
        return BlitzerConf{}, ConfigurationError{fmt.Sprintf("No triggers specified in '%s'",
            filepath.Join(filepath.Dir(yamlPath), "triggers.d"),
        )}
    }
    for _, triggerYamlPath := range triggerPaths {
        ef, err := os.Open(triggerYamlPath)
        if err != nil { return BlitzerConf{}, err }
        yamlBytes, err = ioutil.ReadAll(ef)
        if err != nil { return BlitzerConf{}, err }
        triggerDef := new(TriggerDef)
        goyaml.Unmarshal(yamlBytes, triggerDef)
        if err != nil { return BlitzerConf{}, err }
        conf.TriggerDefs = append(conf.TriggerDefs, triggerDef)

        if conf.Debug == "yes" {
            log.Printf("Loaded trigger configuration from '%s'\n", triggerYamlPath)
        }
    }

    // Load anything probes.d/* files that describe individual probes.
    probePaths, err := filepath.Glob(filepath.Join(filepath.Dir(yamlPath), "probes.d", "*.yaml"))
    if err != nil { return BlitzerConf{}, err }
    if len(probePaths) == 0 {
        return BlitzerConf{}, ConfigurationError{fmt.Sprintf("No probes specified in '%s'",
            filepath.Join(filepath.Dir(yamlPath), "probes.d"),
        )}
    }
    for _, probeYamlPath := range probePaths {
        ef, err := os.Open(probeYamlPath)
        if err != nil { return BlitzerConf{}, err }
        yamlBytes, err = ioutil.ReadAll(ef)
        if err != nil { return BlitzerConf{}, err }
        probeConf := new(ProbeDef)
        goyaml.Unmarshal(yamlBytes, probeConf)
        if err != nil { return BlitzerConf{}, err }
        conf.ProbeDefs = append(conf.ProbeDefs, probeConf)

        if conf.Debug == "yes" {
            log.Printf("Loaded probe configuration from '%s'\n", probeYamlPath)
        }
    }

    return *conf, nil
}
