package main

import (
    "os"
    "path"
    "bytes"
    "errors"
    "net/url"
    "net/http"
    "io/ioutil"
    "encoding/json"

    c "megatask/common"
)

const (
    registerPath = "register"
    loginPath = "login"
    logoutPath = "logout"
    searchPath = "search"
    idPath = "idsearch"
)

const cookieFileName = "mdb_cookie"

func sendGet(url string, cookie *http.Cookie) ([]byte, []*http.Cookie, error) {
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
        return nil, nil, errors.New(string(body))
    }

    cookies := resp.Cookies()
    return body, cookies, err
}

func sendPost(js []byte, address *url.URL) ([]*http.Cookie, error) {
    resp, err := http.Post(address.String(), "application/json", bytes.NewBuffer(js))

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New(string(body))
    }

    cookies := resp.Cookies()
    return cookies, err
}

func storeCookie(cookie *http.Cookie) error {
    js, err := json.Marshal(cookie)

    if err != nil {
        return err
    }

    c.LogD.Printf("Store cookie: %v\n", cookie)
    err = ioutil.WriteFile(cookieFileName, js, 0644)
    return err
}

func readStoredCookie() (*http.Cookie, error) {
    if _, err := os.Stat(cookieFileName); os.IsNotExist(err) {
        return nil, errors.New("Login first !")
    }

    content, err := ioutil.ReadFile(cookieFileName)

    if err != nil {
        return nil, err
    }

    var cookie http.Cookie
    err = json.Unmarshal(content, &cookie)

    return &cookie, err
}

func searchReq(title string, address url.URL, cookie *http.Cookie) error {
    address.Path = path.Join(searchPath)
    q := address.Query()
    q.Add("title", title)
    address.RawQuery = q.Encode()
    rawResp, _, err := sendGet(address.String(), cookie)

    if err != nil {
        return err
    }

    return parseSearchReq(rawResp)
}

func idReq(id string, address url.URL, cookie *http.Cookie) error {
    address.Path = path.Join(idPath)
    q := address.Query()
    q.Add("id", id)
    address.RawQuery = q.Encode()
    rawResp, _, err := sendGet(address.String(), cookie)

    if err != nil {
        return err
    }

    return parseIdReq(rawResp)
}

func logoutReq(address url.URL, cookie *http.Cookie) (*http.Cookie, error) {
    address.Path = path.Join(logoutPath)
    rawResp, cookies, err := sendGet(address.String(), cookie)

    if len(cookies) != 1 {
        return nil, errors.New("Should be only 1 cookie")
    }

    if err != nil {
        return nil, err
    }

    c.LogD.Printf(string(rawResp))
    return cookies[0], nil
}
