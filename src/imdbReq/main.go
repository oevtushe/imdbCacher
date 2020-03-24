package imdbReq

// TODO: use pointers instead raw copy

import (
    "encoding/json"
	"net/http"
    "net/url"
	"io/ioutil"
)

const imdbUrlStr = "https://movie-database-imdb-alternative.p.rapidapi.com/"

type Movie struct {
    Title string
    Year string // not neccessary a number, can be 2008-2012 ...
    ID string `json:"imdbId"`
}

type searchResp struct {
    Search []Movie
    TotalResults int `json:",string"`
    // TODO: how to convert to bool ?
    Response string
}

type MovieExtraInfo struct {
    Title string
    Year string
    Genre string
    Actors string
    Country string
    ImdbRating string
    Production string
    Runtime string
    Type string
    Response string
}

type MovieExtraInfoResp struct {
    MovieExtraInfo
    Response string
}

type ImdbResponseError struct {
    detail string
}

func NewImdbResponseError(detail string) *ImdbResponseError {
    defaultMsg := "ImdbResponseError"
    if detail == "" {
        return &ImdbResponseError{defaultMsg}
    }
    return &ImdbResponseError{detail}
}

func (ire ImdbResponseError) Error() string {
    return ire.detail
}

// TODO: i need to read about everything going on in this func
// TODO: host and key should be read from program parameter
func sendReqInternal(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

    if err != nil {
        return nil, err
    }

	req.Header.Add("x-rapidapi-host", "movie-database-imdb-alternative.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", "3016f5ce2cmsha1f4d33ab43f83fp157df1jsnc54affa5361e")

	res, err := http.DefaultClient.Do(req)

    if err != nil {
        return nil, err
    }

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

    if err != nil {
        return nil, err
    }

    if res.StatusCode != http.StatusOK {
        return nil, NewImdbResponseError(string(body))
    }

    return body, err
}

func parseSearchReq(data []byte) ([]Movie, error) {
    var sr searchResp
    err := json.Unmarshal(data, &sr)

    if err != nil {
        return nil, err
    }

    if sr.Response == "False" {
        return nil, NewImdbResponseError("")
    }

    return sr.Search, err
}

func parseIdReq(data []byte) (*MovieExtraInfoResp, error) {
    var ir MovieExtraInfoResp
    err := json.Unmarshal(data, &ir)

    if err != nil {
        return nil, err
    }

    if ir.Response == "False" {
        return nil, &ImdbResponseError{}
    }

    return &ir, err
}

func SendSearchReq(searchStr string) ([]Movie, error) {
    imdbUrl, err := url.Parse(imdbUrlStr)

    if err != nil {
        return nil, err
    }

    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("s", searchStr)
    imdbUrl.RawQuery = q.Encode()
    var rawResp []byte
    rawResp, err = sendReqInternal(imdbUrl.String())

    if err != nil {
        return nil, err
    }

    return parseSearchReq(rawResp)
}

func SendIdReq(id string) (*MovieExtraInfo, error) {
    imdbUrl, err := url.Parse(imdbUrlStr)

    if err != nil {
        return nil, err
    }

    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("i", id)
    imdbUrl.RawQuery = q.Encode()
    var rawResp []byte
    rawResp, err = sendReqInternal(imdbUrl.String())

    if err != nil {
        return nil, err
    }

    var res *MovieExtraInfoResp
    res, err = parseIdReq(rawResp)

    if err != nil {
        return nil, err
    }

    cp := res.MovieExtraInfo

    return &cp, err
}
