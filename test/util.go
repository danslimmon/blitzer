package blitzertest

import (
    "fmt"
    "regexp"
    "os"
    "path/filepath"
    "io/ioutil"
    "encoding/json"
    "net/http/httptest"
)

type AssertionError struct { Message string }
func (e AssertionError) Error() string { return e.Message }

type ResponseExpectation struct {
    Code int
    BodyRegex string
    BodyJSON string
}

func (re *ResponseExpectation) AssertMatch(recorder *httptest.ResponseRecorder) error {
    err := re.assertMatchCode(recorder.Code)
    if err != nil { return err }

    bodyBytes, err := ioutil.ReadAll(recorder.Body)
    if err != nil { return err }
    bodyStr := string(bodyBytes)
    err = re.assertMatchBodyRegex(bodyStr)
    if err != nil { return err }
    err = re.assertMatchBodyJSON(bodyStr)
    if err != nil { return err }

    return nil
}

func (re *ResponseExpectation) assertMatchCode(code int) error {
    if re.Code == 0 {
        return nil
    }

    if code != re.Code {
        return AssertionError{
            Message: fmt.Sprintf("Expected HTTP code %d; got %d instead", re.Code, code),
        }
    }
    return nil
}

func (re *ResponseExpectation) assertMatchBodyRegex(bodyStr string) error {
    if re.BodyRegex == "" {
        return nil
    }

    if matched, err := regexp.MatchString(re.BodyRegex, bodyStr); ! matched {
        return AssertionError{
            Message: fmt.Sprintf("Body expected to match regex '%s'; got '%s' instead", re.BodyRegex, bodyStr),
        }
    } else if err != nil {
        return AssertionError{
            Message: fmt.Sprintf("Error compiling regex '%s': %s", re.BodyRegex, err),
        }
    }
    return nil
}

func (re *ResponseExpectation) assertMatchBodyJSON(bodyStr string) error {
    if re.BodyJSON == "" {
        return nil
    }

    expBodyMap := make(map[string]interface{}, 0)
    err := json.Unmarshal([]byte(re.BodyJSON), &expBodyMap)
    if err != nil {
        return AssertionError{
            Message: fmt.Sprintf("Error parsing expected JSON '%s': %s", re.BodyJSON, err),
        }
    }

    foundBodyMap := make(map[string]interface{}, 0)
    err = json.Unmarshal([]byte(bodyStr), &foundBodyMap)
    if err != nil {
        return AssertionError{
            Message: fmt.Sprintf("Error parsing body JSON '%s': %s", bodyStr, err),
        }
    }

    for k, expV := range expBodyMap {
        if v, exist := foundBodyMap[k]; ! exist {
            return AssertionError{
                Message: fmt.Sprintf("Body JSON expected to have key '%s' but it is missing", k),
            }
        } else if v != expV {
            return AssertionError{
                Message: fmt.Sprintf("Body JSON expected to have [%s] = '%s', but it is '%s'", k, expV, v),
            }
        }
    }

    return nil
}

// If we're inside a test, make sure to get out of the test dir so that
// templates can be found
func ChdirToMain() {
    Cwd, _ := filepath.Abs(".")
    if filepath.Base(Cwd) == "test" {
        os.Chdir("..")
    }
}
