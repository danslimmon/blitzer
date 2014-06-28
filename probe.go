package main

import (
    "fmt"
    "strings"
    "os"
    "os/exec"
    "io/ioutil"
    "encoding/json"
    "launchpad.net/goyaml"
)

type ProbeResult struct {
    Ref *ProbeRef
    Success bool
    Output string
}

type ProbeRun interface {
    Kickoff() error
}

type AnsibleProbeRun struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult

    playbookPath string
}

func (run *AnsibleProbeRun) Kickoff() error {
    Df("Kicking off Ansible probe '%s' from file '%s'", run.Def.Name, run.Ref.SourceFile)
    playbookPath, err := run.createTempPlaybook()
    if err != nil { return err }
    run.playbookPath = playbookPath
    extra := run.extraVarsStr()
    cmdArgs := []string{
        "-vv",
        "-i",
        Config.Ansible.Inventory,
        "--extra-vars",
        extra,
        run.playbookPath,
    }

    Df("Executing ansible command: ansible-playbook %s\n", strings.Join(cmdArgs, " "))
    cmd := exec.Command("ansible-playbook", cmdArgs...)
    go run.execAsync(cmd)
    return nil
}

func (run *AnsibleProbeRun) createTempPlaybook() (string, error) {
    f, err := ioutil.TempFile("", "blitzer_playbook_")
    defer f.Close()

    pbMap := []map[string]interface{}{
        {
            "hosts": strings.Join([]string{"~", run.Ref.Args["host_pattern"]}, ""),
            "sudo": false,
            "tasks": run.Def.Ansible.Tasks,
        },
    }
    pbBytes, _ := goyaml.Marshal(pbMap)
    _, err = f.Write(pbBytes)
    if err != nil { return "", err }
    return f.Name(), nil
}

func (run *AnsibleProbeRun) deleteTempPlaybook() error {
    return os.Remove(run.playbookPath)
}

func (run *AnsibleProbeRun) extraVarsStr() string {
    rslt, _ := json.Marshal(run.Ref.Args)
    return string(rslt)
}

func (run *AnsibleProbeRun) makeResult(procErr error, outputBytes []byte) *ProbeResult {
    rslt := new(ProbeResult)
    rslt.Ref = run.Ref
    rslt.Success = true

    if procErr != nil {
        rslt.Success = false
    }
    rslt.Output = string(outputBytes)
    Df("Received output from probe '%s':\n===== BEGIN OUTPUT =====\n%s\n===== END OUTPUT =====\n", rslt.Ref.Name, rslt.Output)
    return rslt
}

func (run *AnsibleProbeRun) execAsync(cmd *exec.Cmd) {
    output, err := cmd.CombinedOutput()
    rslt := run.makeResult(err, output)
    run.RsltChan <- rslt
    run.deleteTempPlaybook()
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
