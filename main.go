package main

import (
    "os"
    "log"
)

var Config *BlitzerConf

// Logs a Println-type message if we're in debug mode
func D(s string) {
    if Config.Debug == "yes" {
        log.Println(s)
    }
}

// Logs a Printf-type message if we're in debug mode
func Df(format string, v ...interface{}) {
    if Config.Debug == "yes" {
        log.Printf(format, v...)
    }
}

func main() {
    c, err := GetConf("etc/blitzer.yaml")
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
    Config = &c
}
