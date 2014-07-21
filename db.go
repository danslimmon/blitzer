package main

import (
    "strings"
    "encoding/json"
    "github.com/fzzy/radix/redis"
)

var dbInst *DB

type DB struct {
    RedisClient *redis.Client
}

func (db *DB) historyRedisKey(incSlug string) string {
    return strings.Join([]string{"history", incSlug}, "_")
}

func (db *DB) WriteHistory(inc *Incident, he HistoryEvent) error {
    prJSON, err := json.Marshal(he)
    if err != nil { return err }
    _, err = db.RedisClient.Cmd("zadd", db.historyRedisKey(inc.Slug), he.Timestamp, string(prJSON)).Int()
    if err != nil { return err }
    return nil
}

// Returns all events for the given incident since the given time.
//
// Events are returned in chronological order.
func (db *DB) HistoryEventsSince(incSlug string, timestamp int64) ([]HistoryEvent, error) {
    rslt := make([]HistoryEvent, 0)
    blobs, err := db.RedisClient.Cmd("zrangebyscore", db.historyRedisKey(incSlug), timestamp, 1<<62).ListBytes()
    if err != nil { return rslt, err }

    for _, j := range blobs {
        he := new(HistoryEvent)
        err := json.Unmarshal(j, he)
        if err != nil { return make([]HistoryEvent, 0), err }
        rslt = append(rslt, *he)
    }
    return rslt, nil
}

func GetDB() (*DB, error) {
    if dbInst == nil {
        c, err := redis.Dial("tcp", "127.0.0.1:6379")
        if err != nil { return &DB{}, err }
        return &DB{RedisClient: c}, nil
    }
    return dbInst, nil
}
