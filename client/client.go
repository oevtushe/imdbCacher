package main

import (
    "fmt"
    "flag"
    "path"
    "strconv"
    "net/url"
    "io/ioutil"
    "encoding/json"

    c "imdbCacher/common"
)

func parseSearchReq(data []byte) error {
    var movies []c.Movie
    err := json.Unmarshal(data, &movies)

    if err != nil {
        return err
    }

    for _, movie := range movies {
        fmt.Printf("---------------\n")
        fmt.Printf("Title: %v\n", movie.Title)
        fmt.Printf("Year: %v\n", movie.Year)
        fmt.Printf("id: %v\n", movie.ID)
        fmt.Printf("---------------\n")
    }
    return nil
}

func parseIdReq(data []byte) error {
    var movie c.MovieExtraInfo
    err := json.Unmarshal(data, &movie)

    if err != nil {
        return err
    }

    fmt.Printf("---------------\n")
    fmt.Printf("Title: %v\n", movie.Title)
    fmt.Printf("Year: %v\n", movie.Year)
    fmt.Printf("Genre: %v\n", movie.Genre)
    fmt.Printf("Actors: %v\n", movie.Actors)
    fmt.Printf("Runtime: %v\n", movie.Runtime)
    fmt.Printf("Country: %v\n", movie.Country)
    fmt.Printf("Rating: %v\n", movie.ImdbRating)
    fmt.Printf("Type: %v\n", movie.Type)
    fmt.Printf("---------------\n")
    return nil
}


func search(fsearch string, mservUrl *url.URL) {
    cookie, err := readStoredCookie()

    if err != nil {
        c.LogE.Fatal(err)
    }

    err = searchReq(fsearch, *mservUrl, cookie)

    if err != nil {
        c.LogE.Fatal(err)
    }
}

func idSearch(fId string, mservUrl *url.URL) {
    cookie, err := readStoredCookie()

    if err != nil {
        c.LogE.Fatal(err)
    }

    err = idReq(fId, *mservUrl, cookie)

    if err != nil {
        c.LogE.Fatal(err)
    }

}

func login(name, pass string, mservUrl url.URL) {
    c.LogD.Printf("user=%v, pass=%v\n", name, pass)
    // TODO: hash password
    creds, err := json.Marshal(c.User{name, pass})

    if err != nil {
        c.LogE.Fatal(err)
    }

    mservUrl.Path = path.Join(loginPath)
    cookies, err := sendPost(creds, &mservUrl)

    if err != nil {
        c.LogE.Fatal(err)
    }

    if len(cookies) != 1 {
        c.LogE.Fatal("Should be only 1 cookie")
    }

    fmt.Printf("Login successful !\n")
    err = storeCookie(cookies[0])

    if err != nil {
        c.LogE.Fatal(err)
    }

    c.LogD.Printf("Cookie stored\n")
}

func register(name, pass string, mservUrl url.URL) {
    c.LogD.Printf("user=%v, pass=%v\n", name, pass)
    creds, err := json.Marshal(c.User{name, pass})

    if err != nil {
        c.LogE.Fatal(err)
    }

    mservUrl.Path = path.Join(registerPath)
    _, err = sendPost(creds, &mservUrl)

    if err != nil {
        c.LogE.Fatal(err)
    }

    fmt.Printf("Registration successful !\n")
}

func logout(mservUrl *url.URL) {
    cookie, err := readStoredCookie()

    if err != nil {
        c.LogE.Fatal(err)
    }

    resCookie, err := logoutReq(*mservUrl, cookie)

    if err != nil {
        c.LogE.Fatal(err)
    }

    if err := storeCookie(resCookie); err != nil {
        c.LogE.Fatal(err)
    }

    fmt.Printf("Logout successful !\n")
}

func usage() {
    fmt.Printf("usage: go run imdbCacher/client -login|register " +
            "-user <username> -password <pass> [-port <num>]\n" +
            "       go run imdbCacher/client -search <movie title> [-port <num>]\n" +
            "       go run imdbCacher/client -id <movie id> [-port <num>]\n")
    flag.PrintDefaults()
}

func main() {
    c.LogD.SetOutput(ioutil.Discard)
    fId := flag.String("id", "", "search movie by imdb id returned after -search request")
    fSearch := flag.String("search", "", "search movie by title in imdb")
    fUser := flag.String("user", "", "user name")
    fPass := flag.String("password", "", "user password")
    fRegister := flag.Bool("register", false, "user registeration")
    fLogin := flag.Bool("login", false, "login operation")
    fLogout := flag.Bool("logout", false, "logout operation")
    fPort := flag.Int("port", 8080, "port the server is listening on")
    flag.Parse()
    address := "http://127.0.0.1:" + strconv.Itoa(*fPort)
    mservUrl, err := url.Parse(address)

    if err != nil {
        c.LogE.Fatal(err)
    }

    c.LogD.Printf("Server address: %v\n", address)

    // TODO: temporary quick solution
    if *fSearch != "" {
        search(*fSearch, mservUrl)
    } else if *fId != "" {
        idSearch(*fId, mservUrl)
    } else if *fLogin && *fUser != "" && *fPass != "" {
        login(*fUser, *fPass, *mservUrl)
    } else if *fRegister && *fUser != "" && *fPass != "" {
        register(*fUser, *fPass, *mservUrl)
    } else if *fLogout {
        logout(mservUrl)
    } else {
        usage()
        return
    }
}
