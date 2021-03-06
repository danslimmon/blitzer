package blitzertest

import (
    "testing"
    "strings"
    "net/http"
    "net/http/httptest"
    "github.com/zenazn/goji"
    blitzer "github.com/danslimmon/blitzer"
)

func Test_POST_Event_Nagios(t *testing.T) {
    blitzer.PopulateControllers()
    blitzer.PopulateConfOrBarf("test/etc/blitzer.yaml")

    // Normal behavior (200 OK with URL of new incident)
    reqbody := strings.NewReader(`{"service":"foo","state":"CRITICAL"}`)
    resp := httptest.NewRecorder()
    req, err := http.NewRequest("POST", "/event/nagios", reqbody)
    if err != nil { t.Fatal(err) }

    exp := ResponseExpectation{Code: 200, BodyRegex: `^{"slug":".*"}$`}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }

    // Send an incomplete event object (400 Bad Request)
    reqbody = strings.NewReader(`{"state":"CRITICAL"}`)
    resp = httptest.NewRecorder()
    req, err = http.NewRequest("POST", "/event/nagios", reqbody)
    if err != nil { t.Fatal(err) }

    exp = ResponseExpectation{Code: 400, BodyJSON: `{"error":"Missing 'service' parameter"}`}

    goji.DefaultMux.ServeHTTP(resp, req)
    if err = exp.AssertMatch(resp); err != nil {
        t.Fatal(err)
    }
}
