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

func Test_GET_Incident_IncidentSlug_History_Timestamp(t *testing.T) {
    blitzer.PopulateControllers()
    ChdirToMain()
    log.Println(filepath.Abs("."))

    db, err := blitzer.GetDB()
    if err != nil { t.Fatal(err) }
    err = db.WriteHistory(
        &blitzer.Incident{Slug: "test_incident_please_ignore"},
        blitzer.HistoryEvent{
            Timestamp: 1234567890,
            Success: true,
            ProbeName: "null_probe",
            ProbeType: "graphite",
            Values: make(map[string]string, 0),
        },
    )
    if err != nil { t.Fatal(err) }
    err = db.WriteHistory(
        &blitzer.Incident{Slug: "test_incident_please_ignore"},
        blitzer.HistoryEvent{
            Timestamp: 1234567893,
            Success: true,
            ProbeName: "null_probe",
            ProbeType: "graphite",
            Values: make(map[string]string, 0),
        },
    )
    if err != nil { t.Fatal(err) }

    // Full listing (200 OK)
    resp := httptest.NewRecorder()
    req, err := http.NewRequest("GET", "/incident/test_incident_please_ignore/history/0", nil)
    if err != nil { t.Fatal(err) }

    exp := ResponseExpectation{Code: 200, BodyJSON: `{"result":[{"timestamp":1234567893,"success":true,"probe_name":"null_probe","probe_type":"graphite","values":{}},{"timestamp":1234567890,"success":true,"probe_name":"null_probe","probe_type":"graphite","values":{}}]}`}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }

    // Partial listing (200 OK)
    resp = httptest.NewRecorder()
    req, err = http.NewRequest("GET", "/incident/test_incident_please_ignore/history/1234567891", nil)
    if err != nil { t.Fatal(err) }

    exp = ResponseExpectation{Code: 200, BodyJSON: `{"result":[{"timestamp":1234567893,"success":true,"probe_name":"null_probe","probe_type":"graphite","values":{}}]}`}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }

    // Empty listing (200 OK)
    resp = httptest.NewRecorder()
    req, err = http.NewRequest("GET", "/incident/test_incident_please_ignore/history/1234567894", nil)
    if err != nil { t.Fatal(err) }

    exp = ResponseExpectation{Code: 200, BodyJSON: `{"result":[]}`}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }

}
