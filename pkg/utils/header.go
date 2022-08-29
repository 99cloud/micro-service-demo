package utils

import (
	"net/http"
)

func CopyHeader(dest, src http.Header) {
	if dest == nil || src == nil {
		return
	}
	for key, vals := range src {
		for _, val := range vals {
			dest.Set(key, val)
		}
	}
}
