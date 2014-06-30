package main

import (
    "fmt"
    "net/http"
    "github.com/zenazn/goji/web"
)

func GET_Incident(c web.C, w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", c.URLParams["incident_slug"])
}
