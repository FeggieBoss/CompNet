package internals

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"log"
	"os"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func Locate(url string, resp Response) string {
	fileName := fmt.Sprint(hash(url))
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	f.WriteString(resp.Status + "\n")
	f.WriteString(resp.Url + "\n")
	f.WriteString(resp.LastModified + "\n")
	f.WriteString(resp.Etag + "\n")
	f.WriteString(resp.Body + "\n")

	return fileName
}

func Read(fileName string) Response {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	status := scanner.Text()

	scanner.Scan()
	url := scanner.Text()

	scanner.Scan()
	lastModified := scanner.Text()

	scanner.Scan()
	etag := scanner.Text()

	body := ""
	for scanner.Scan() {
		body += scanner.Text()
	}

	return Response{
		Status:       status,
		Url:          url,
		LastModified: lastModified,
		Etag:         etag,
		Body:         body,
	}
}
