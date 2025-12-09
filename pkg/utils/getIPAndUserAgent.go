package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIPString(r *http.Request) string {
	// Get the X-Forwarded-For header
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For may contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0]) // Return first IP after trimming spaces
	}

	// Fallback to RemoteAddr if X-Forwarded-For is empty
	ip, _, _ = net.SplitHostPort(r.RemoteAddr) // Remove port if present
	return ip
}

// GetUserAgent extracts the User-Agent from the request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
