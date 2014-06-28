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
    Error error
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

// Converts the output from an ansible-playbook run to user-showable data 
//
// Normally, we just take the output of all 'shell' commands executed and
// concatenate their stdout. If any command returned nonzero, its output
// is replaced with an error message.
func (run *AnsibleProbeRun) parseAnsibleOutput(output string) string {
    type cmdRslt struct {
        Cmd string
        RC int8
        Stdout string
        Stderr string
    }
    type rsltPiece struct {
        Hostname string
        CmdRslt *cmdRslt
    }

    lines := strings.Split(output, "\n")
    pieces := make([]rsltPiece, 0)
    for _, line := range lines {
        if strings.Index(line, "changed: [") != 0 { continue }
        piece := rsltPiece{Hostname: "", CmdRslt: new(cmdRslt)}
        piece.Hostname = line[strings.Index(line, "[")+1:strings.Index(line, "]")]
        jsonBytes := []byte(line[strings.Index(line, "{"):])
        if err := json.Unmarshal(jsonBytes, piece.CmdRslt); err != nil {
            return fmt.Sprintf("Invalid JSON in Ansible output line: %s\n", line)
        }
        pieces = append(pieces, piece)
    }

    rsltLines := make([]string, 0)
    for _, piece := range pieces {
        rsltLines = append(rsltLines, fmt.Sprintf("%s ~~~~~~~~~~~~~~~~~~~~", piece.Hostname))
        if piece.CmdRslt.RC == 0 {
            rsltLines = append(rsltLines, piece.CmdRslt.Stdout)
        } else {
            rsltLines = append(rsltLines, fmt.Sprintf("<Command exited with return code %d and stderr as follows>\n%s", piece.CmdRslt.Stderr))
        }
    }

    return strings.Join(rsltLines, "\n")
}

func (run *AnsibleProbeRun) makeResult(procErr error, outputBytes []byte) *ProbeResult {
    rslt := new(ProbeResult)
    rslt.Ref = run.Ref
    rslt.Success = true
    if procErr != nil {
        rslt.Success = false
        rslt.Error = procErr
        return rslt
    }

    output := run.parseAnsibleOutput(string(outputBytes))
    rslt.Output = output

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
