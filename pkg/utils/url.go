package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseUrl(u string) (*url.URL, error) {
	e := u
	if !strings.HasPrefix(e, "http") {
		e = fmt.Sprintf("http://%s", u)
	}
	return url.Parse(e)
}
