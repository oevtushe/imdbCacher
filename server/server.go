package main

import (
    "fmt"
    "log"
    "time"
    "flag"
    "strconv"
    "net/http"
    "io/ioutil"
    "database/sql"
    "encoding/json"

    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"

    "megatask/imdb"
    c "megatask/common"
)

type Movie struct {
    c.Movie
    Info string
}

// DB entry lifetime in minutes
const dbDataLifetime = 15
const cookieName = "cookie"
// lifetime in minutes
const cookieLifetime = 15 * 60

var dba *DBAccessor
var store *sessions.CookieStore

func serverUsage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Usage: /logout, /login, /register, /search, /idsearch\n")
}

func search(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        loginRequired(w)
        return
    }

    values := r.URL.Query()
    title := values.Get("title")

    if title == "" {
        http.Error(w, "Invalid search request, " +
                "you must specify \"title\"", http.StatusBadRequest)
        return
    }

    c.LogD.Printf("User (%v) requests (%v)",
            session.Values["user"], title)
    sr, err := imdb.SendSearchReq(title)

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

func idSearch(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        loginRequired(w)
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
        http.Error(w, "Invalid cookie\n", http.StatusInternalServerError)
        return
    }

    c.LogD.Printf("User (%v) requests movie with id (%v)", user, id)

    js, err := getMovie(id, user)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func register(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var user c.User
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

func logout(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, cookieName)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !checkIfLogin(session) {
        // TODO: stop repeating yourself
        loginRequired(w)
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

    c.LogD.Printf("User (%v) logged out !\n", user)
    fmt.Fprintf(w, "Logout successful\n")
}

func login(w http.ResponseWriter, r *http.Request) {
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

    var user c.User
    err = json.Unmarshal(body, &user)

    if err != nil {
        http.Error(w, "Wrong content format: " + err.Error(),
                http.StatusBadRequest)
        return
    }

    c.LogD.Printf("POST data: %v\n", user)
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

    c.LogD.Printf("User (%v) logged in !\n", userCred.Login)
    fmt.Fprintf(w, "Login successful\n")
}

// load movie from imdb api or local db
func getMovie(id string, user string) ([]byte, error) {
    movie, err := dba.GetMovie(id)
    var cached bool

    if err != nil {
        if err != sql.ErrNoRows {
            return nil, err
        }
    } else {
        cached = true
    }

    c.LogD.Printf("IsMovieCached: %v\n", cached)
    if !cached {
        ir, err := imdb.SendIdReq(id)

        if err != nil {
            return nil, err
        }

        js, err := json.Marshal(ir)

        if err != nil {
            return nil, err
        }

        movie = &Movie{
            Movie: ir.Movie,
            Info: string(js),
        }
        err = dba.AddMovie(user, *movie,
                time.Now().Add(time.Minute * dbDataLifetime))

        if err != nil {
            return nil, err
        }
    }

    return []byte(movie.Info), nil
}

func checkIfLogin(s *sessions.Session) bool {
    u, ok := s.Values["user"]
    return ok == true && u != ""
}

func loginRequired(w http.ResponseWriter) {
    http.Error(w, "Login first", http.StatusUnauthorized)
}

func init() {

    authKeyOne := securecookie.GenerateRandomKey(64)
    encryptionKeyOne := securecookie.GenerateRandomKey(32)
    store = sessions.NewCookieStore(authKeyOne, encryptionKeyOne)
    store.Options = &sessions.Options {
        MaxAge: cookieLifetime,
        HttpOnly: true,
    }
}

func execUsage() {
    fmt.Printf("go run megatask/server -key <key> -user <name> " +
            "-password <pass> -dbname <name> [-port <num>]\n")
    flag.PrintDefaults()
}

func initDb(name, password, dbname string) {
    var err error
    dbInfo := fmt.Sprintf("user=%v password=%v dbname=%v", name, password, dbname)
    dba, err = NewDBAcessor("postgres", dbInfo)

    if err != nil {
        c.LogE.Fatal(err)
    }

    dba.CreateTablesIfNeeded()

    if err != nil {
        c.LogE.Fatal(err)
    }

}

func main() {
    key := flag.String("key", "", "(x-rapidapi-key) IMDB Alternative key")
    port := flag.Int("port", 8080, "port to start server on")
    user := flag.String("user", "", "db username")
    password := flag.String("password", "", "db user password")
    dbname := flag.String("dbname", "", "db name")
    flag.Parse()

    if *key != "" && *user != "" && *password != "" && *dbname != "" {
        imdb.InitImdb(*key)
        initDb(*user, *password, *dbname)
    } else {
        execUsage()
        return
    }

    router := mux.NewRouter()
    router.HandleFunc("/", serverUsage)
    router.HandleFunc("/login", login)
    router.HandleFunc("/logout", logout)
    router.HandleFunc("/register", register)
    router.HandleFunc("/search", search)
    router.HandleFunc("/idsearch", idSearch)
    addr := ":" + strconv.Itoa(*port)
    c.LogD.Printf("Start on: %v\n", addr)
    log.Fatal(http.ListenAndServe(addr, router))
}
