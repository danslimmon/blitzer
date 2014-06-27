package main

import (
    "fmt"
    "log"
    "os"
    "io"
    "sort"
    "path/filepath"
    "io/ioutil"
    "crypto/md5"
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

// A reference to a ProbeDef, as included in a trigger.
type ProbeRef struct {
    Name string
    Args map[string]string
}
// Deterministically hashes Name and Args to a unique string
func (pr *ProbeRef) Hash() string {
    h := md5.New()
    argKeys := make([]string, 0)
    for k, _ := range pr.Args { argKeys = append(argKeys, k) }
    sort.Strings(argKeys)

    io.WriteString(h, pr.Name)
    for _, k := range argKeys {
        y, _ := goyaml.Marshal(map[string]string{k:pr.Args[k]})
        io.WriteString(h, string(y))
    }

    return(fmt.Sprintf("%x", h.Sum(nil)))
}

type TriggerDef struct {
    ServiceMatch string "service_match"
    ProbeRefs []*ProbeRef
}

type AnsibleProbeDef struct {
    Module string
    Args string
}

type ProbeDef struct {
    Name string
    Title string
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
