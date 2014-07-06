package blitzer

import (
    "fmt"
    "log"
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
    _, err := NewEventFromNagios(r)
    if err != nil {
        BarfJSON(c, w, r, err)
    } else {
        w.WriteHeader(204)
    }
}

func GET_Incident_IncidentSlug(c web.C, w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("./views/incident_incidentslug.html")
    if err != nil {
        BarfJSON(c, w, r, err)
        return
    }
    w.WriteHeader(200)
    err = tmpl.Execute(w, nil)
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
