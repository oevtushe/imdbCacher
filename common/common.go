package common

import (
    "os"
    "log"
)

type Movie struct {
    Title string
    Year string // not neccessary a number, can be 2008-2012 ...
    ID string `json:"imdbId"`
}

type MovieExtraInfo struct {
    Movie
    Genre string
    Actors string
    Country string
    ImdbRating string
    Production string
    Runtime string
    Type string
}

type User struct {
    Login, Pass string
}

var LogE *log.Logger = log.New(os.Stdout, "[ERROR]: ", 0)
var LogD *log.Logger = log.New(os.Stdout, "[DEBUG]: ", log.Lshortfile)
