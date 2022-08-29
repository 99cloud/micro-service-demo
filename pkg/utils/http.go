package utils

import (
	"net/http"
	"strconv"
)

func HTTPStatusCode(w http.ResponseWriter, r *http.Request) {
	code := http.StatusInternalServerError
	if c, err := strconv.Atoi(r.URL.Query().Get("code")); err == nil &&
		c > 99 && c < 600 {
		code = c
	}
	w.WriteHeader(code)
	_, _ = w.Write([]byte(http.StatusText(code)))
}
