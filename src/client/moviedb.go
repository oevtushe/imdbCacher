package main

import (
    "encoding/json"
    "flag"
    "net/http"
    "strconv"
    "io/ioutil"
    "os"
    "fmt"
    "net/url"
    "log"
    "path"
    "bytes"
)

var logD *log.Logger = log.New(os.Stdout, "[DEBUG]: ", log.Lshortfile)

const cookieFileName = "mdb_cookie"

type User struct {
    Login string
    Pass string
}

type ServerError struct {
    detail string
}

type ClientError struct {
    detail string
}

func NewClientError(detail string) *ServerError {
    defaultMsg := "ImdbResponseError"
    if detail == "" {
        return &ServerError{defaultMsg}
    }
    return &ServerError{detail}
}

const (
    registerPath = "register"
    loginPath = "login"
    logoutPath = "logout"
    searchPath = "search"
    idPath = "idsearch"
)

func NewServerError(detail string) *ServerError {
    defaultMsg := "ImdbResponseError"
    if detail == "" {
        return &ServerError{defaultMsg}
    }
    return &ServerError{detail}
}

func (se ServerError) Error() string {
    return se.detail
}

func (cl ClientError) Error() string {
    return cl.detail
}

func sendGetReqInternal(url string, cookie *http.Cookie) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

    if err != nil {
        return nil, err
    }

    req.AddCookie(cookie)
    resp, err := http.DefaultClient.Do(req)

    if err != nil {
        return nil, err
    }

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, NewServerError(string(body))
    }

    return body, err
}

func sendGetWithCookieReqInternal(url string, cookie *http.Cookie) ([]byte, *http.Cookie, error) {
	req, err := http.NewRequest("GET", url, nil)

    if err != nil {
        return nil, nil, err
    }

    req.AddCookie(cookie)
    resp, err := http.DefaultClient.Do(req)

    if err != nil {
        return nil, nil, err
    }

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        return nil, nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, nil, NewServerError(string(body))
    }

    cookies := resp.Cookies()

    if len(cookies) != 1 {
        // TODO: return err
        logD.Fatal("Should be only 1 cookie")
    }

    return body, cookies[0], err
}

func storeCookie(cookie *http.Cookie) error {
    js, err := json.Marshal(cookie)

    if err != nil {
        return err
    }

    logD.Printf("Store cookie: %v\n", cookie)
    err = ioutil.WriteFile(cookieFileName, js, 0644)
    return err
}

func readStoredCookie() (*http.Cookie, error) {
    if _, err := os.Stat(cookieFileName); os.IsNotExist(err) {
        return nil, NewClientError("Login first !")
    }

    content, err := ioutil.ReadFile(cookieFileName)

    if err != nil {
        return nil, err
    }

    var cookie http.Cookie
    err = json.Unmarshal(content, &cookie)

    return &cookie, err
}

func parseSearchReq(data []byte) error {
    fmt.Printf("Result: %v\n", string(data))
    return nil
}

func SendSearchReq(title string, address url.URL, cookie *http.Cookie) error {
    address.Path = path.Join(searchPath)
    q := address.Query()
    q.Add("title", title)
    address.RawQuery = q.Encode()
    rawResp, err := sendGetReqInternal(address.String(), cookie)

    if err != nil {
        return err
    }

    return parseSearchReq(rawResp)
}

func parseIdReq(data []byte) error {
    fmt.Printf("Result: %v\n", string(data))
    return nil
}

func SendIdReq(id string, address url.URL, cookie *http.Cookie) error {
    address.Path = path.Join(idPath)
    q := address.Query()
    q.Add("id", id)
    address.RawQuery = q.Encode()
    rawResp, err := sendGetReqInternal(address.String(), cookie)

    if err != nil {
        return err
    }

    return parseIdReq(rawResp)
}

func SendLoginReq(creds []byte, address url.URL) (*http.Cookie, error) {
    address.Path = path.Join(loginPath)

    resp, err := http.Post(address.String(), "application/json", bytes.NewBuffer(creds))

    if err != nil {
        return nil, err
    }

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, NewServerError(string(body))
    }

    cookies := resp.Cookies()

    if len(cookies) != 1 {
        // TODO: return err
        logD.Fatal("Should be only 1 cookie")
    }

    return cookies[0], err
}

func SendRegisterReq(creds []byte, address url.URL) error {
    address.Path = path.Join(registerPath)

    resp, err := http.Post(address.String(), "application/json", bytes.NewBuffer(creds))

    if err != nil {
        return err
    }

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        return err
    }

    if resp.StatusCode != http.StatusOK {
        return NewServerError(string(body))
    }

    return err
}

func SendLogoutReq(address url.URL, cookie *http.Cookie) (*http.Cookie, error) {
    address.Path = path.Join(logoutPath)
    rawResp, resCookie, err := sendGetWithCookieReqInternal(address.String(), cookie)

    if err != nil {
        return nil, err
    }

    logD.Printf(string(rawResp))

    return resCookie, nil
}

func main() {
    search := flag.String("search", "", "search movie by title in imdb")
    id := flag.String("id", "", "search movie by imdb id returned after -search request")
    user := flag.String("user", "", "user name")
    pass := flag.String("password", "", "user password")
    register := flag.Bool("register", false, "user registeration")
    login := flag.Bool("login", false, "login operation")
    logout := flag.Bool("logout", false, "logout operation")
    port := flag.Int("port", 8080, "port the server is listening on")

    flag.Parse()

    //var cookie *http.Cookie
    address := "http://127.0.0.1:" + strconv.Itoa(*port)
    mservUrl, err := url.Parse(address)

    if err != nil {
        logD.Fatal(err)
    }

    logD.Printf("Address: %v\n", address)

    // TODO: escape string
    if *search != "" {
        cookie, err := readStoredCookie()

        if err != nil {
            logD.Fatal(err)
        }

        err = SendSearchReq(*search, *mservUrl, cookie)

        if err != nil {
            logD.Fatal(err)
        }

    } else if *id != "" {
        cookie, err := readStoredCookie()

        if err != nil {
            logD.Fatal(err)
        }

        err = SendIdReq(*id, *mservUrl, cookie)

        if err != nil {
            logD.Fatal(err)
        }

    } else if *login && *user != "" && *pass != "" {
        logD.Printf("user=%v, pass=%v\n", *user, *pass)
        // TODO: hash password first please !
        creds, err := json.Marshal(User{*user, *pass})

        if err != nil {
            logD.Fatal(err)
        }

        cookie, err := SendLoginReq(creds, *mservUrl)

        if err != nil {
            logD.Fatal(err)
        }

        // TODO: :/
        fmt.Printf("Login successful !\n")
        err = storeCookie(cookie)

        if err != nil {
            logD.Fatal(err)
        }

        fmt.Printf("Cookie is stored\n")
    } else if *register && *user != "" && *pass != "" {
        logD.Printf("user=%v, pass=%v\n", *user, *pass)
        creds, err := json.Marshal(User{*user, *pass})

        if err != nil {
            logD.Fatal(err)
        }

        err = SendRegisterReq(creds, *mservUrl)

        if err != nil {
            logD.Fatal(err)
        }

        fmt.Printf("Registration successful !\n")

    } else if *logout {
        cookie, err := readStoredCookie()

        if err != nil {
            logD.Fatal(err)
        }

        resCookie, err := SendLogoutReq(*mservUrl, cookie)

        if err != nil {
            if err, ok := err.(*ServerError); ok {
                logD.Fatalf("ServerError: %v", err)
            }
            logD.Fatal(err)
        }

        if err := storeCookie(resCookie); err != nil {
            logD.Fatal(err)
        }

        fmt.Printf("Logout successful !\n")
    }
}
