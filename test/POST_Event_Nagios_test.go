package blitzertest

import (
    "testing"
    "strings"
    "net/http"
    "net/http/httptest"
    "github.com/zenazn/goji"
    "github.com/danslimmon/blitzer"
)

func Test_POST_Event_Nagios(t *testing.T) {
    blitzer.PopulateControllers()

    // Normal behavior (204 No Content)
    reqbody := strings.NewReader(`{"service":"foo","state":"CRITICAL"}`)
    resp := httptest.NewRecorder()
    req, err := http.NewRequest("POST", "/event/nagios", reqbody)
    if err != nil { t.Fatal(err) }

    exp := ResponseExpectation{Code: 204, BodyRegex: "^$"}

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
