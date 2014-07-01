package blitzer

import (
    "fmt"
    "strings"
    "os"
    "os/exec"
    "io/ioutil"
    "encoding/json"
    "launchpad.net/goyaml"
)

type AnsibleConf struct {
    Inventory string
}

type AnsibleProbe struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult

    playbookPath string
}

func (p *AnsibleProbe) Init(ref *ProbeRef, def *ProbeDef, ch chan *ProbeResult) {
    p.Ref = ref
    p.Def = def
    p.RsltChan = ch
}

func (p *AnsibleProbe) Kickoff() error {
    Df("Kicking off Ansible probe '%s' from file '%s'", p.Def.Name, p.Ref.SourceFile)
    playbookPath, err := p.createTempPlaybook()
    if err != nil { return err }
    p.playbookPath = playbookPath
    ctx := ProbeContext(p.Def, p.Ref)
    extra := p.extraVarsStr(ctx)
    cmdArgs := []string{
        "-vv",
        "-i",
        Config.Ansible.Inventory,
        "--extra-vars",
        extra,
        p.playbookPath,
    }

    Df("Executing ansible command: ansible-playbook %s\n", strings.Join(cmdArgs, " "))
    cmd := exec.Command("ansible-playbook", cmdArgs...)
    go p.execAsync(cmd)
    return nil
}

func (p *AnsibleProbe) createTempPlaybook() (string, error) {
    f, err := ioutil.TempFile("", "blitzer_playbook_")
    defer f.Close()

    pbMap := []map[string]interface{}{
        {
            "hosts": strings.Join([]string{"~", p.Ref.Args["host_pattern"]}, ""),
            "sudo": false,
            "tasks": p.Def.Args["tasks"],
        },
    }
    pbBytes, _ := goyaml.Marshal(pbMap)
    _, err = f.Write(pbBytes)
    if err != nil { return "", err }
    return f.Name(), nil
}

func (p *AnsibleProbe) deleteTempPlaybook() error {
    return os.Remove(p.playbookPath)
}

func (p *AnsibleProbe) extraVarsStr(ctx map[string]interface{}) string {
    rslt, _ := json.Marshal(ctx)
    return string(rslt)
}

// Converts the output from an ansible-playbook p to user-showable data 
//
// Normally, we just take the output of all 'shell' commands executed and
// concatenate their stdout. If any command returned nonzero, its output
// is replaced with an error message.
func (p *AnsibleProbe) parseAnsibleOutput(output string) string {
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

func (p *AnsibleProbe) makeResult(procErr error, outputBytes []byte) *ProbeResult {
    rslt := new(ProbeResult)
    rslt.Ref = p.Ref
    rslt.Success = true
    if procErr != nil {
        rslt.Success = false
        rslt.Error = procErr
        return rslt
    }

    rslt.Values = make(map[string]string, 0)
    output := p.parseAnsibleOutput(string(outputBytes))
    rslt.Values["Output"] = output

    return rslt
}

func (p *AnsibleProbe) execAsync(cmd *exec.Cmd) {
    output, err := cmd.CombinedOutput()
    rslt := p.makeResult(err, output)
    p.RsltChan <- rslt
    p.deleteTempPlaybook()
}
