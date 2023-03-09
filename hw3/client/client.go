package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalln("Bad arguments")
	}
	serv := os.Args[1] + ":" + os.Args[2]
	conn, err := net.Dial("tcp", serv)
	if err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go copyTo(os.Stdout, conn, &wg)

	req := http.Request{Method: "GET", URL: &url.URL{Host: serv, Path: "/" + os.Args[3]}}
	data, err := httputil.DumpRequest(&req, false)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}

	wg.Wait()
}

func copyTo(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatalln(err)
	}
}
