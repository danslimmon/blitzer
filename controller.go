package main

import (
    "fmt"
    "log"
    "strconv"
    "net/http"
    "encoding/json"
    "html/template"
    "github.com/zenazn/goji/web"
)

type WebError struct {
    Code int
    Message string
    E error
}
func (e WebError) Error() string {
    if e.E != nil {
        return e.E.Error()
    } else {
        return e.Message
    }
}

func POST_Event_Nagios(c web.C, w http.ResponseWriter, r *http.Request) {
    ev, err := NewEventFromNagios(r)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }

    tds, err := MatchTriggerDefs(ev)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }

    inc, err := NewIncident(ev, tds)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    log.Println(inc.State)
    w.WriteHeader(204)
}

func GET_Incident_IncidentSlug(c web.C, w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("./views/incident_incidentslug.html")
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    w.WriteHeader(200)
    err = tmpl.Execute(w, nil)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
}

// AJAX endpoint returning recent history events for the given incident
//
// URL is like /incident/2014-07-21_incidentname/history/1234567890
//
// This would return all history events since the timestamp 1234567890 inclusive,
// in a format like this:
//
//    {"result": [
//        {"timestamp":1234567890","success":"true",probe_name:"whatever","values":{...}},
//        {"timestamp":1234567894", ...},
//        ...
//    ]}
func GET_Incident_IncidentSlug_History_Timestamp(c web.C, w http.ResponseWriter, r *http.Request) {
    incSlug := c.URLParams["incident_slug"]
    timestamp, err := strconv.ParseInt(c.URLParams["timestamp"], 10, 64)
    if err != nil {
        BarfJSON(c, w, r, WebError{Code: 400, Message: "Invalid timestamp after /history/"})
        return
    }

    rslt := make(map[string][]HistoryEvent, 0)
    db, err := GetDB()
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    events, err := db.HistoryEventsSince(incSlug, timestamp)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    rslt["result"] = events

    j, err := json.Marshal(rslt)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }

    w.WriteHeader(200)
    w.Write(j)
}

func errorJSON(e error) string {
    m := make(map[string]string, 0)
    m["error"] = e.Error()
    j, err := json.Marshal(m)
    if err != nil {
        log.Println("Error parsing error JSON:", e)
        return `{"error":"Internal Server Error"}`
    }
    return string(j)
}

func BarfJSON(c web.C, w http.ResponseWriter, r *http.Request, e error) {
    we, ok := e.(WebError)
    if ok {
        w.WriteHeader(we.Code)
        log.Println("Error:", we)
        fmt.Fprintf(w, errorJSON(we))
    } else {
        w.WriteHeader(500)
        log.Println("Error:", e)
        fmt.Fprintf(w, `{"error":"Internal Server Error"}`)
    }
}
