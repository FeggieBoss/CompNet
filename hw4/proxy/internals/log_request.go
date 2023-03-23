package internals

import (
	"fmt"
	"log"
	"os"
)

func LogRequest(url string, status string) {
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("url: %s, status: %s\n", url, status)); err != nil {
		log.Fatalln(err)
	}
}
