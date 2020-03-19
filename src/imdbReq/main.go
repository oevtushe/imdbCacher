package imdbReq

// TODO: use pointers instead raw copy

import (
	"fmt"
    "encoding/json"
	"net/http"
    "net/url"
	"io/ioutil"
)

type Search struct {
    Poster string
    Title string
    Year int `json:",string"`
    ImdbID string
    Type string
}

type SearchResp struct {
    Search []Search
    TotalResults int `json:",string"`
    // TODO: how to convert to bool ?
    Response string
}

type IdResp struct {
    Title string
    Genre string
    Actors string
    Country string
    ImdbRating float32 `json:",string"`
    Production string
}

const imdbURL = "https://movie-database-imdb-alternative.p.rapidapi.com/"

// TODO: i need to read about everything going on in this func
func sendReqInternal(url string) []byte {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-host", "movie-database-imdb-alternative.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", "3016f5ce2cmsha1f4d33ab43f83fp157df1jsnc54affa5361e")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
    // TODO: error handling
	body, err := ioutil.ReadAll(res.Body)

    if err != nil {
        panic(err)
    }

    return body
}

func parseSearchReq(data []byte) SearchResp {
    var sr SearchResp
    err := json.Unmarshal(data, &sr)

    if err != nil {
        panic(err)
    }

    return sr
}

func parseIdReq(data []byte) IdResp {
    var ir IdResp
    err := json.Unmarshal(data, &ir)

    if err != nil {
        panic(err)
    }

    return ir
}

const imdbUrlStr = "https://movie-database-imdb-alternative.p.rapidapi.com/"

func SendSearchReq(searchStr string) SearchResp {
    imdbUrl,_ := url.Parse(imdbUrlStr)
    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("s", searchStr)
    imdbUrl.RawQuery = q.Encode()
    rawResp := sendReqInternal(imdbUrl.String())
    return parseSearchReq(rawResp)
}

func SendIdReq(id string) IdResp {
    imdbUrl,_ := url.Parse(imdbUrlStr)
    q := imdbUrl.Query()
    q.Add("r", "json")
    q.Add("i", id)
    imdbUrl.RawQuery = q.Encode()
    rawResp := sendReqInternal(imdbUrl.String())
    return parseIdReq(rawResp)
}
