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

type BlitzerConf struct {
    Addr string
    Port int16
    Debug string
    Ansible AnsibleConf
    Graphite GraphiteConf
    TriggerDefs []*TriggerDef
    ProbeDefs []*ProbeDef
}

func (c *BlitzerConf) Validate() error {
    if len(c.TriggerDefs) < 1 {
        return ConfigurationError{"No triggers defined in triggers.d"}
    }
    if len(c.ProbeDefs) < 1 {
        return ConfigurationError{"No triggers defined in probes.d"}
    }

    for _, td := range c.TriggerDefs {
        err := td.Validate()
        if err != nil { return err }
    }

    for _, pd := range c.ProbeDefs {
        err := pd.Validate()
        if err != nil { return err }
    }

    return nil
}


// A reference to a ProbeDef, as included in a trigger.
type ProbeRef struct {
    Name string
    Args map[string]string
    SourceFile string
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
    ProbeRefs []*ProbeRef "probes"
    SourceFile string
}

func (td *TriggerDef) Validate() error {
    if td.ServiceMatch == "" {
        return ConfigurationError{fmt.Sprintf("No service_match specified for trigger in '%s'", td.SourceFile)}
    }
    if len(td.ProbeRefs) == 0 {
        return ConfigurationError{fmt.Sprintf("No probes specified for trigger in '%s'", td.SourceFile)}
    }
    return nil
}

type AnsibleProbeDef struct {
    Tasks []map[string]string
}

type GraphiteProbeDef struct {
    QSTemplate string "qs_template"
}

type ProbeDef struct {
    Name string
    Title string
    Type string
    Html string
    SourceFile string
    Interval int64 "interval_ms"

    // Types of probe we support
    Ansible AnsibleProbeDef
    Graphite GraphiteProbeDef
}

func (pd *ProbeDef) Validate() error {
    if pd.Name == "" {
        return ConfigurationError{fmt.Sprintf("No name specified for probe in '%s'", pd.SourceFile)}
    }
    if pd.Type == "" {
        return ConfigurationError{fmt.Sprintf("No type specified for probe in '%s'", pd.SourceFile)}
    }
    if pd.Interval == 0 {
        return ConfigurationError{fmt.Sprintf("No interval_ms specified for probe in '%s'", pd.SourceFile)}
    }
    return nil
}

func (pd *ProbeDef) PopulateDefaults() {
    if pd.Title == "" {
        pd.Title = pd.Name
    }
    if pd.Html == "" {
        pd.Html = "{{.}}"
    }
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

        triggerDef.SourceFile = triggerYamlPath
        for _, pr := range triggerDef.ProbeRefs {
            pr.SourceFile = triggerYamlPath
        }

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
        probeDef := new(ProbeDef)
        goyaml.Unmarshal(yamlBytes, probeDef)
        if err != nil { return BlitzerConf{}, err }
        probeDef.SourceFile = probeYamlPath
        probeDef.PopulateDefaults()
        conf.ProbeDefs = append(conf.ProbeDefs, probeDef)

        if conf.Debug == "yes" {
            log.Printf("Loaded probe configuration from '%s'\n", probeYamlPath)
        }
    }

    return *conf, nil
}

func PopulateConfOrBarf(yamlPath string) {
    c, err := GetConf(yamlPath)
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    err = c.Validate()
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    Config = &c
}
