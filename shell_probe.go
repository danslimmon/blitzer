package main

import (
    "fmt"
    "time"
    "strings"
    "os/exec"
    "text/template"
)

type ShellConf struct {
    Inventory string
}

type ShellProbe struct {
    Ref *ProbeRef
    Def *ProbeDef
    RsltChan chan *ProbeResult
}

func (p *ShellProbe) Init(ref *ProbeRef, def *ProbeDef, ch chan *ProbeResult) {
    p.Ref = ref
    p.Def = def
    p.RsltChan = ch
}

func (p *ShellProbe) Kickoff() error {
    Df("Kicking off shell probe '%s' from file '%s'", p.Def.Name, p.Ref.SourceFile)
    cmdStr, ok := p.Def.Args["command"].(string)
    if ! ok { return ConfigurationError{fmt.Sprintf("Shell probe's 'command' arg must be a string in '%s'", p.Def.SourceFile)} }
    cmdSlice := strings.Split(cmdStr, " ")

    Df("Executing shell command: %s\n", cmdStr)
    cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
    go p.execAsync(cmd)
    return nil
}

func (p *ShellProbe) evalTemplate(tplStr string) (string, error) {
    tmpl, err := template.New("shell_command").Parse(tplStr)
    writer := &stringWriter{}
    ctx := ProbeContext(p.Def, p.Ref)
    err = tmpl.Execute(writer, ctx)
    return writer.S, err
}

func (p *ShellProbe) execAsync(cmd *exec.Cmd) {
    output, err := cmd.CombinedOutput()
    rslt := p.makeResult(err, output)
    p.RsltChan <- rslt
}

func (p *ShellProbe) makeResult(procErr error, outputBytes []byte) *ProbeResult {
    rslt := new(ProbeResult)
    rslt.Ref = p.Ref
    rslt.Success = true
    rslt.Timestamp = time.Now().Unix()
    rslt.Values = make(map[string]string, 0)
    output := string(outputBytes)
    rslt.Values["output"] = output

    if procErr != nil {
        rslt.Success = false
        rslt.Error = procErr
        return rslt
    }

    return rslt
}
