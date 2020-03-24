package main

import (
    "net/http"
    "fmt"
    "log"
    "os"
    "time"
    "io/ioutil"
    "encoding/json"
    "database/sql"

    _ "github.com/lib/pq"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"

    "imdbReq"
)

// TODO: improve error handling
// TODO: handle no internet connection
// TODO: pass objects by pointers

// TODO: meh
type User struct {
    Login, Pass string
}

type Movie struct {
    Title string
    Year string // not neccessary a number, can be 2008-2012 ...
    Info string
    ID string
}

var dba *DBAccessor
var store *sessions.CookieStore
const cookieName = "cookie"
// DB entry lifetime in minutes
const dbDataLifetime = 15

var logW *log.Logger = log.New(os.Stdout, "[WARNING]: ", log.Lshortfile)
var logE *log.Logger = log.New(os.Stdout, "[ERROR]: ", log.Lshortfile)
var logD *log.Logger = log.New(os.Stdout, "[DEBUG]: ", log.Lshortfile)

// TODO: bueautify
func handleUsage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Usage: /login, /register, /search, /idsearch\n")
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

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
                "you must specify \"title\"", http.StatusBadRequest)
        return
    }

    // TODO: session.Values["user"] -> user better to be some constant
    logD.Printf("User (%v) requests (%v)",
            session.Values["user"], title)

    sr, err := imdbReq.SendSearchReq(title)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    js, err := json.Marshal(sr)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func handleIdSearch(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        http.Error(w, "Login first", http.StatusUnauthorized)
        return
    }

    values := r.URL.Query()
    id := values.Get("id")

    if id == "" {
        http.Error(w, "Invalid search request, " +
                "you must specify \"id\"", http.StatusBadRequest)
        return
    }

    user, ok := session.Values["user"].(string)

    if !ok {
        logE.Printf("Invalid session cookie\n")
        http.Error(w, "Invalid cookie\n", http.StatusInternalServerError)
        return
    }

    logD.Printf("User (%v) requests movie with id (%v)", user, id)

    movie, err := dba.GetMovie(id)
    var cached bool

    if err != nil {
        if err != sql.ErrNoRows {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    } else {
        cached = true
    }

    var js []byte

    if !cached {
        var ir *imdbReq.MovieExtraInfo
        // TODO: what if no such movie in imdb ?
        ir, err = imdbReq.SendIdReq(id)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        js, err = json.Marshal(ir)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        movie = &Movie {
            ID: id,
            Title: ir.Title,
            Year: ir.Year,
            Info: string(js),
        }

        err = dba.AddMovie(user, *movie, time.Now().Add(time.Minute * dbDataLifetime))

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    js, err = json.Marshal(movie)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func checkIfLogin(s *sessions.Session) bool {
    u, ok := s.Values["user"]
    return ok == true && u != ""
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var user User
    err = json.Unmarshal(body, &user)

    if err != nil {
        http.Error(w, "Wrong content format: " + err.Error(),
                http.StatusBadRequest)
        return
    }

    _, err = dba.GetUser(user.Login)

    if err == nil {
        http.Error(w, "Username is already taken", http.StatusConflict)
        return
    } else if err != sql.ErrNoRows {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := dba.AddUser(user); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Registration successful\n")
}

// TODO: server track login/logout only by presence of cookie
func handleLogout(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        // TODO: stop repeating yourself
        http.Error(w, "Login first", http.StatusUnauthorized)
        return
    }

    user := session.Values["user"]
    session.Values["user"] = ""
    session.Options.MaxAge = -1

    err = session.Save(r, w)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    logD.Printf("User (%v) logged out !\n", user)
    fmt.Fprintf(w, "Logout successful\n")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var user User
    err = json.Unmarshal(body, &user)

    // TODO: too much repeating
    if err != nil {
        http.Error(w, "Wrong content format: " + err.Error(),
                http.StatusBadRequest)
        return
    }

    logD.Printf("POST data: %v\n", user)
    userCred, err := dba.GetUser(user.Login)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "No such user, register first",
                    http.StatusBadRequest)
            return
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    if userCred.Pass != user.Pass {
        http.Error(w, "Wrong password !", http.StatusBadRequest)
        return
    }

    session.Values["user"] = user.Login
    err = session.Save(r, w)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    logD.Printf("User (%v) logged in !\n", userCred.Login)
    fmt.Fprintf(w, "Login successful\n")
}

func init() {
    // TODO: do as options for program
    var err error
    dba, err = NewDBAcessor("postgres", "user=test password=pass1 dbname=test")

    if err != nil {
        logE.Fatal(err)
    }

    dba.CreateTables()

    if err != nil {
        logE.Fatal(err)
    }

    authKeyOne := securecookie.GenerateRandomKey(64)
    encryptionKeyOne := securecookie.GenerateRandomKey(32)
    store = sessions.NewCookieStore(authKeyOne, encryptionKeyOne)
    store.Options = &sessions.Options {
        MaxAge: 60 * 15,
        HttpOnly: true,
    }
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/", handleUsage)
    router.HandleFunc("/login", handleLogin)
    router.HandleFunc("/logout", handleLogout)
    router.HandleFunc("/register", handleRegister)
    router.HandleFunc("/search", handleSearch)
    // TODO: is it necessary ?
    router.HandleFunc("/idsearch", handleIdSearch)
    log.Fatal(http.ListenAndServe(":8080", router))
}
