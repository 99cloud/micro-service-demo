package utils

import (
	"net/http"
	"os"
)

func InjectPodName(header http.Header) {
	header.Set("Pod-Name", GetPodName())
}

func GetPodName() string {
	if podName := os.Getenv("POD_NAME"); podName != "" {
		return podName
	}
	hostname, _ := os.Hostname()
	return hostname
}
