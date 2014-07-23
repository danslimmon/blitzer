package main

import (
    "os"
    "fmt"
    "log"
    "io/ioutil"
    "strconv"
    "net/http"
    "encoding/json"
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

    inc, exists := GetIncidentByEvent(ev)
    switch {
    case exists && inc.State == "active" && ev.State == "up":
        // The service is back up
        Df("Deactivating incident '%s' because service is back up", inc.Slug)
        inc.Deactivate()
    case !exists && ev.State == "down":
        // Service is newly down
        Df("Received a new 'down' alert for service '%s'", ev.ServiceName)
        tds, err := MatchTriggerDefs(ev)
        if err != nil {
            BarfJSON(c, w, r, err)
            return
        }
        _, err = NewIncident(ev, tds)
        if err != nil {
            BarfJSON(c, w, r, err)
            return
        }
    }
    w.WriteHeader(204)
}

func GET_Incident_IncidentSlug(c web.C, w http.ResponseWriter, r *http.Request) {
    f, err := os.Open("/Users/dan/blitzer/views/incident_incidentslug.html")
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    w.WriteHeader(200)
    outBytes, err := ioutil.ReadAll(f)
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    w.Write(outBytes)
}

// AJAX endpoint returning recent history events for the given incident
//
// URL is like /incident/2014-07-21_incidentname/history/1234567890
//
// This would return all history events since the timestamp 1234567890 inclusive,
// in reverse chronological order, in a format like this:
//
//    {"result": [
//        {"timestamp":1234567894","success":"true",probe_name:"whatever","values":{...}},
//        {"timestamp":1234567890", ...},
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
