package main

import (
    "os"
    "log"
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
    c, err := GetConf("etc/blitzer.yaml")
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    Config = &c

    inc, err := NewIncident(Config.TriggerDefs[0].ProbeRefs)
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    D(inc.State)
}
