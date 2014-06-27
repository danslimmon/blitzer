package main

import (
    "os"
    "log"
    "time"
)

var Config *BlitzerConf

// Logs a Println-type message if we're in debug mode
func D(v ...interface{}) {
    if Config.Debug == "yes" {
        log.Println(v...)
    }
}

// Logs a Printf-type message if we're in debug mode
func Df(format string, v ...interface{}) {
    if Config.Debug == "yes" {
        log.Printf(format, v...)
    }
}

func main() {
    PopulateConfOrBarf("etc/blitzer.yaml")

    inc, err := NewIncident(&Event{ServiceName:"Search API"}, Config.TriggerDefs[0].ProbeRefs)
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    D(inc.State)
    time.Sleep(5 * time.Millisecond)
    inc.Deactivate()
    time.Sleep(5 * time.Millisecond)
    D(inc.State)
}
