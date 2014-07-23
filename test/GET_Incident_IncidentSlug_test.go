package blitzertest

import (
    "log"
    "path/filepath"

    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/zenazn/goji"
    blitzer "github.com/danslimmon/blitzer"
)

func Test_GET_Incident_IncidentSlug(t *testing.T) {
    blitzer.PopulateControllers()
    ChdirToMain()
    log.Println(filepath.Abs("."))

    // Normal behavior (200 OK)
    resp := httptest.NewRecorder()
    req, err := http.NewRequest("GET", "/incident/test_incident_please_ignore", nil)
    if err != nil { t.Fatal(err) }

    exp := ResponseExpectation{Code: 200, BodyRegex: "ng-app"}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }

}
