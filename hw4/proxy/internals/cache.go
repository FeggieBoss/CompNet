package internals

import (
	"io"
	"log"
	"net/http"
	"strings"
)

type Response struct {
	Status       string
	Url          string
	LastModified string
	Etag         string
	Body         string
}

func (r Response) HTTPresponse() http.Response {
	ret := http.Response{Status: r.Status, Body: io.NopCloser(strings.NewReader(r.Body))}
	ret.Header = make(map[string][]string)
	ret.Header.Add("Last-Modified", r.LastModified)
	ret.Header.Add("Etag", r.Etag)
	return ret
}

func SearchCache(cache map[string]string, url string) (Response, bool) {
	respFileName, ok := cache[url]
	if ok {
		resp := Read(respFileName)

		rqst, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalln(err)
		}

		rqst.Header.Add("If-Modified-Since", resp.LastModified)
		rqst.Header.Add("If-None-Match", resp.Etag)

		client := http.Client{}
		checkResp, err := client.Do(rqst)
		if err != nil {
			log.Fatalln(err)
		}

		if checkResp.Status == "304 Not Modified" {
			return resp, true
		} else {
			return Response{}, false
		}
	} else {
		return Response{}, false
	}
}
