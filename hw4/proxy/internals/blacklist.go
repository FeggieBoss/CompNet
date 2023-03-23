package internals

import (
	"bufio"
	"os"
)

func CheckBlacklist(url string) bool {
	file, err := os.Open("blacklist.txt")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == url {
			return true
		}
	}

	return false
}
