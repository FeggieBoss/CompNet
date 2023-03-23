package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"proxy/internals"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Bad arguments")
	}
	l, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()

	var (
		limitN, _ = strconv.Atoi(os.Args[1])
		stopper   = make(chan int, limitN)
		cache     = make(map[string]string)
	)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		stopper <- 1
		go handleClient(conn, stopper, cache)
	}
}

func handleClient(conn net.Conn, stopper chan int, cache map[string]string) {
	defer func() {
		conn.Close()
		<-stopper
	}()

	var (
		request *http.Request
		err     error
	)
	if request, err = http.ReadRequest(bufio.NewReader(conn)); err != nil {
		log.Fatalln(err)
	}

	var (
		respInternalRaw = doRequest(request, cache)
		respInternal    = respInternalRaw.HTTPresponse()
	)

	resp := http.Response{Status: respInternal.Status}
	if respInternal.Body != nil {
		resp.Body = respInternal.Body
	}

	err = resp.Write(conn)
	if err != nil {
		log.Fatalln(err)
	}
}

func doRequest(request *http.Request, cache map[string]string) internals.Response {
	var (
		body = request.Body
		url  = request.URL.String()
	)

	if internals.CheckBlacklist(url) {
		return internals.Response{Status: "400 Bad Request", Url: url, Body: "The requested page is blacklisted"}
	}

	respBodyMsg, ok := internals.SearchCache(cache, url)
	if ok {
		return respBodyMsg
	}

	if request.Method == "GET" {
		body = nil
	}

	var rqstInternal, err = http.NewRequest(request.Method, url, body)
	if err != nil {
		log.Fatalln(err)
	}

	client := http.Client{}
	resp, err := client.Do(rqstInternal)
	if err != nil {
		log.Fatalln(err)
	}

	internals.LogRequest(url, resp.Status)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	respRaw := internals.Response{
		Status: resp.Status,
		Url:    url,
		Body:   string(raw),
	}

	if lm := resp.Header.Get("Last-Modified"); len(lm) != 0 {
		respRaw.LastModified = lm
	}
	if e := resp.Header.Get("Etag"); len(e) != 0 {
		respRaw.Etag = e
	}

	cache[url] = internals.Locate(url, respRaw)

	return respRaw
}
