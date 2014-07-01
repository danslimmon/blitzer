package blitzer

import (
    "regexp"
)

// Determines which TriggerDefs in our configuration match the given event
func MatchTriggerDefs(event Event) ([]*TriggerDef, error) {
    rslt := make([]*TriggerDef, 0)
    for _, td := range Config.TriggerDefs {
        matched, err := regexp.MatchString(td.ServiceMatch, event.ServiceName)
        if err != nil { return make([]*TriggerDef, 0), err }
        if matched {
            rslt = append(rslt, td)
        }
    }
    return rslt, nil
}
