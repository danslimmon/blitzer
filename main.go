package main

import (
    "log"
    "net/http"
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

func PopulateControllers() error {
    goji.Get("/static/*", http.FileServer(http.Dir("/Users/dan/blitzer")))
    goji.Get("/incident/:incident_slug", GET_Incident_IncidentSlug)
    goji.Get("/incident/:incident_slug/history/:timestamp", GET_Incident_IncidentSlug_History_Timestamp)

    goji.Post("/event/nagios", POST_Event_Nagios)

    return nil
}

func main() {
    PopulateConfOrBarf("etc/blitzer.yaml")
    err := PopulateControllers()
    if err != nil { log.Fatal(err) }

    goji.Serve()
}
