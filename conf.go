package main

import (
    "fmt"
    //@DEBUG
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
    Probes []*ProbeConf
}

type AnsibleProbeDef struct {
    Module string
    Args string
}

type ProbeConf struct {
    Name string
    MatchHost string
    MatchService string
    Html string
    Ansible AnsibleProbeDef
}

// Loads our configuration from the given path
func LoadConf(yamlPath string) (BlitzerConf, error) {
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
        probeConf := new(ProbeConf)
        goyaml.Unmarshal(yamlBytes, probeConf)
        if err != nil { return BlitzerConf{}, err }
        conf.Probes = append(conf.Probes, probeConf)

        if conf.Debug == "yes" {
            log.Printf("Loaded probe configuration from '%s'\n", probeYamlPath)
        }
    }

    return *conf, nil
}
