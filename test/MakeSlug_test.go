package blitzertest

import (
    "testing"
    "regexp"
    blitzer "github.com/danslimmon/blitzer"
)

func Test_MakeSlug(t *testing.T) {
    t.Parallel()

    e := &blitzer.Event{
        ServiceName: "Foo Blah Service",
        State: "down",
    }
    re := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}_foo_blah_service")
    slug := blitzer.MakeSlug(e)
    if nil == re.Find([]byte(slug)) {
        t.Fatal("Slug didn't match expected format:", slug)
    }

    e = &blitzer.Event{
        ServiceName: "Foo ---27 //whatever?",
        State: "down",
    }
    re = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}_foo_27_whatever")
    slug = blitzer.MakeSlug(e)
    if nil == re.Find([]byte(slug)) {
        t.Fatal("Slug didn't match expected format:", slug)
    }
}
