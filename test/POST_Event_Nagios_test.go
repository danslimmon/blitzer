package blitzertest

import (
    "testing"
    "strings"
    "io/ioutil"
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

    goji.DefaultMux.ServeHTTP(resp, req)
    if _, err = ioutil.ReadAll(resp.Body); err != nil {
        t.Fatal("Failed to read response")
    } else {
        if resp.Code != 204 {
            t.Fatalf("Incorrect status code %d", resp.Code)
        }
    }

    // Send an incomplete event object (400 Bad Request)
    reqbody = strings.NewReader(`{"state":"CRITICAL"}`)
    resp = httptest.NewRecorder()
    req, err = http.NewRequest("POST", "/event/nagios", reqbody)
    if err != nil { t.Fatal(err) }

    goji.DefaultMux.ServeHTTP(resp, req)
    if _, err = ioutil.ReadAll(resp.Body); err != nil {
        t.Fatal("Failed to read response")
    } else {
        if resp.Code != 400 {
            t.Fatalf("Incorrect status code %d", resp.Code)
        }
    }

}
