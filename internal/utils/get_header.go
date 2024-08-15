package utils

import (
	"fmt"
	"net/http"
	"strings"
)

func GetHeaders(r *http.Request) string {
	var headers []string
	for name, values := range r.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", name, strings.Join(values, ", ")))
	}
	return strings.Join(headers, "\n")
}
