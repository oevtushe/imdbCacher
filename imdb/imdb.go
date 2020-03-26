package imdb

// TODO: use pointers instead raw copy

import (
    "errors"
    "net/url"
	"net/http"
	"io/ioutil"
    "encoding/json"

    "megatask/common"
)

var xRapidApiKey string
const imdbUrlStr = "https://movie-database-imdb-alternative.p.rapidapi.com/"

type idResp struct {
    common.MovieExtraInfo
    Response string
}

type searchResp struct {
    Search []common.Movie
    TotalResults int `json:",string"`
    Response string
}

func InitImdb(key string) {
    xRapidApiKey = key
}

// TODO: host and key should be read as program parameter
func sendReq(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

    if err != nil {
        return nil, err
    }

	req.Header.Add("x-rapidapi-host", "movie-database-imdb-alternative.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", xRapidApiKey)

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
        return nil, errors.New(string(body))
    }

    return body, err
}

func parseSearchReq(data []byte) ([]common.Movie, error) {
    var sr searchResp
    err := json.Unmarshal(data, &sr)

    if err != nil {
        return nil, err
    }

    if sr.Response == "False" {
        return nil, errors.New("Movie not found")
    }

    return sr.Search, err
}

func parseIdReq(data []byte) (*idResp, error) {
    var ir idResp
    err := json.Unmarshal(data, &ir)

    if err != nil {
        return nil, err
    }

    if ir.Response == "False" {
        return nil, errors.New("Movie not found")
    }

    return &ir, err
}

func SendSearchReq(searchStr string) ([]common.Movie, error) {
    imdbUrl, err := url.Parse(imdbUrlStr)

    if err != nil {
        return nil, err
    }

    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("s", searchStr)
    imdbUrl.RawQuery = q.Encode()
    rawResp, err := sendReq(imdbUrl.String())

    if err != nil {
        return nil, err
    }

    return parseSearchReq(rawResp)
}

func SendIdReq(id string) (*common.MovieExtraInfo, error) {
    imdbUrl, err := url.Parse(imdbUrlStr)

    if err != nil {
        return nil, err
    }

    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("i", id)
    imdbUrl.RawQuery = q.Encode()
    rawResp, err := sendReq(imdbUrl.String())

    if err != nil {
        return nil, err
    }

    res, err := parseIdReq(rawResp)

    if err != nil {
        return nil, err
    }

    cp := res.MovieExtraInfo

    return &cp, err
}
