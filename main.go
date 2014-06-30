package main

import (
    "log"
    "github.com/zenazn/goji"
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
    goji.Get("/:incident_slug", GET_Incident)
    goji.Serve()
}
