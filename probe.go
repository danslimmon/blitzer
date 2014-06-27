package main

import (
    "fmt"
    "strings"
    "io"
    "os/exec"
    "io/ioutil"
    "encoding/json"
)

type ProbeResult struct {
    Ref *ProbeRef
    Output string
}

type ProbeRun interface {
    Kickoff() error
}

type AnsibleProbeRun struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult
}

func (run *AnsibleProbeRun) Kickoff() error {
    Df("Kicking off Ansible probe '%s' from file '%s'", run.Def.Name, run.Ref.SourceFile)
    playbookPath, err := run.createTempPlaybook()
    if err != nil { return err }
    extra := run.extraVarsStr()
    cmdArgs := []string{
        "-i",
        Config.Ansible.Inventory,
        "--extra-vars",
        extra,
        playbookPath,
    }

    Df("Executing ansible command: ansible-playbook %s\n", strings.Join(cmdArgs, " "))
    _ = exec.Command("ansible-playbook", cmdArgs...)
    return nil
}

func (run *AnsibleProbeRun) createTempPlaybook() (string, error) {
    f, err := ioutil.TempFile("", "blitzer_playbook_")
    defer f.Close()
    _, err = io.WriteString(f, fmt.Sprintf(`---
- hosts: ~%s
  tasks:
    - %s: "%s"`, run.Ref.Args["host_pattern"], run.Def.Ansible.Module, run.Def.Ansible.Args))
    if err != nil { return "", err }
    return f.Name(), nil
}

func (run *AnsibleProbeRun) extraVarsStr() string {
    rslt, _ := json.Marshal(run.Ref.Args)
    return string(rslt)
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
