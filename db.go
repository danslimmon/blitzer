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

func (db *DB) historyRedisKey(inc *Incident) string {
    return strings.Join([]string{"history", inc.Slug}, "_")
}

func (db *DB) WriteHistory(inc *Incident, he HistoryEvent) error {
    prJSON, err := json.Marshal(he)
    if err != nil { return err }
    _, err = db.RedisClient.Cmd("zadd", db.historyRedisKey(inc), he.Timestamp, string(prJSON)).Int()
    if err != nil { return err }
    return nil
}

func getDB() (*DB, error) {
    if dbInst == nil {
        c, err := redis.Dial("tcp", "127.0.0.1:6379")
        if err != nil { return &DB{}, err }
        return &DB{RedisClient: c}, nil
    }
    return dbInst, nil
}
