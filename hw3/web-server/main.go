package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const CRLF = "\r\n"

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Bad arguments")
	}
	l, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()

	limitN, _ := strconv.Atoi(os.Args[1])
	stopper := make(chan int, limitN)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		stopper <- 1
		go handleClient(conn, stopper)
	}
}

func handleClient(conn net.Conn, stopper chan int) {
	defer conn.Close()

	buf := make([]byte, 5000)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}

	fileExists := false

	requestLine := string(buf[:n])

	tokens := strings.Split(requestLine, " ")
	fileName := tokens[1]

	data, err := os.ReadFile(string(fileName[1:]))
	if err == nil {
		fileExists = true
	}

	var resp http.Response
	if fileExists {
		resp = http.Response{Status: "200 OK", Body: io.NopCloser(bytes.NewReader(data))}
	} else {
		resp = http.Response{Status: "404 NOT FOUND", Body: io.NopCloser(strings.NewReader("File not found!"))}
	}

	err = resp.Write(conn)
	if err != nil {
		log.Fatalln(err)
	}

	<-stopper
}
