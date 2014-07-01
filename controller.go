package blitzer

import (
    "fmt"
    "log"
    "net/http"
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
        Barf(c, w, r, err)
    } else {
        w.WriteHeader(204)
    }
}

func GET_IncidentSlug(c web.C, w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", c.URLParams["incident_slug"])
}

func Barf(c web.C, w http.ResponseWriter, r *http.Request, e error) {
    we, ok := e.(WebError)
    if ok {
        w.WriteHeader(we.Code)
        log.Println("Error:", we)
        fmt.Fprintf(w, "Error %d: %s\n\n%s", we.Code, http.StatusText(we.Code), we)
    } else {
        w.WriteHeader(500)
        log.Println("Error:", e)
        fmt.Fprintf(w, "Internal Server Error", http.StatusText(500), e)
    }
}
