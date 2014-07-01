package blitzer

import (
    "fmt"
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
        Barf(c, w, r, &WebError{E: err})
    } else {
        w.WriteHeader(204)
    }
}

func GET_IncidentSlug(c web.C, w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", c.URLParams["incident_slug"])
}

func Barf(c web.C, w http.ResponseWriter, r *http.Request, e *WebError) {
    fmt.Fprintf(w, "Error %d: %s\n\n%s", e.Code, http.StatusText(e.Code), e)
    http.Error(w, e.Error(), e.Code)
}
