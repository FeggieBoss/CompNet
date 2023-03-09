package main

import (
	"log"
	"net"
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

func contentType(fileName string) string {
	if strings.HasSuffix(fileName, ".htm") || strings.HasSuffix(fileName, ".html") {
		return "text/html"
	}

	if strings.HasSuffix(fileName, ".ram") || strings.HasSuffix(fileName, ".ra") {
		return "audio/x-pn-realaudio"
	}

	return "application/octet-stream"
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

	statusLine := ""
	contentTypeLine := ""
	entityBody := ""
	if fileExists {
		statusLine = "HTTP/1.0 200 OK" + CRLF
		contentTypeLine = "Content-Type: " + contentType(fileName) + CRLF
	} else {
		statusLine = "HTTP/1.0 404 Not Found" + CRLF
		contentTypeLine = "Content-Type: text/html" + CRLF
		entityBody = "<HTML>" +
			"<HEAD><TITLE>Not Found</TITLE></HEAD>" +
			"<BODY>Not Found</BODY></HTML>"
	}

	var sb strings.Builder
	sb.WriteString(statusLine)
	sb.WriteString(contentTypeLine)
	sb.WriteString(CRLF)

	if fileExists {
		sb.WriteString(string(data))
	} else {
		sb.WriteString(entityBody)
	}

	_, err = conn.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalln(err)
	}

	<-stopper
}
