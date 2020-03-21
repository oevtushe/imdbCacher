package main

import (
    "net/http"
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"

    "imdbReq"
)

// TODO: improve error handling

// TODO: meh
type User struct {
    Login, Pass string
}

var db = make(map[string]string)
var store *sessions.CookieStore

var logW *log.Logger
var logE *log.Logger
var logD *log.Logger

func init() {
    db["sasha"] = "pass1"
    db["anton"] = "pass2"
    db["yevhen"] = "pass3"

    authKeyOne := securecookie.GenerateRandomKey(64)
    encryptionKeyOne := securecookie.GenerateRandomKey(32)
    store = sessions.NewCookieStore(authKeyOne, encryptionKeyOne)
    store.Options = &sessions.Options {
        MaxAge: 60,
        HttpOnly: true,
    }

    logW = log.New(os.Stdout, "[WARNING]: ", log.Lshortfile)
    logE = log.New(os.Stdout, "[ERROR]: ", log.Lshortfile)
    logD = log.New(os.Stdout, "[DEBUG]: ", log.Lshortfile)
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/", handleUsage)
    router.HandleFunc("/login", handleLogin)
    router.HandleFunc("/logout", handleLogout)
    router.HandleFunc("/register", handleRegister)
    router.HandleFunc("/search", handleSearch)
    router.HandleFunc("/idsearch", handleIdSearch)
    log.Fatal(http.ListenAndServe(":8080", router))
}

// TODO: bueautify
func handleUsage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Usage: /login, /register, /search, /idsearch\n")
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "cookie")

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        http.Error(w, "Login first", http.StatusUnauthorized)
        return
    }

    values := r.URL.Query()

    title := values.Get("title")

    if title == "" {
        http.Error(w, "Invalid search request, " +
                "you must specify \"title\"\n", http.StatusBadRequest)
        return
    }

    logD.Printf("Title: %v", title)
    sr := imdbReq.SendSearchReq(title)
    res := ""

    for _, t := range sr.Search {
        res += t.Title + "\n"
    }

    fmt.Fprintf(w, res)
}

func handleIdSearch(w http.ResponseWriter, r *http.Request) {
}

func checkIfLogin(s *sessions.Session) bool {
    _, ok := s.Values["user"]
    return ok == true
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        panic(err)
    }

    var user User
    err = json.Unmarshal(body, &user)

    if err != nil {
        // TODO: err should be shown
        fmt.Fprintf(w, "json.Unmarshal failed\n")
        return
    }

    fmt.Println(user)
    _, ok := db[user.Login]

    if ok == true {
        //http conflict
        fmt.Fprintf(w, "Username is already taken\n")
        return
    }

    db[user.Login] = user.Pass
    fmt.Fprintf(w, "Registration successful\n")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "cookie")

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        // TODO: stop repeating yourself
        http.Error(w, "Login first", http.StatusUnauthorized)
        return
    }

    session.Values["user"] = ""
    session.Options.MaxAge = -1

    err = session.Save(r, w)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Logout successful\n")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "cookie")

    if err != nil {
        // TODO: i use different type of loggers through project ...
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        // TODO: panic may be not so approptiate
        logE.Printf(err.Error())
        return
    }

    var user User
    err = json.Unmarshal(body, &user)

    if err != nil {
        // TODO: panic may be not so approptiate
        logE.Printf(err.Error())
        return
    }

    _, ok := db[user.Login]

    if ok == true {
        session.Values["user"] = user.Login
        err = session.Save(r, w)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Login successful\n")
    }

    fmt.Println(string(body))
}
