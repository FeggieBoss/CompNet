package internals

import "strings"

func ConvertDomain(sURL string) string {
	if strings.Contains(sURL, ":") {
		return sURL
	}
	return sURL + ":80"
}
