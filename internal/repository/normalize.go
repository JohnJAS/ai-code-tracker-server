package repository

import (
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var scpStyleRemote = regexp.MustCompile(`^(?:[^@/\s]+@)?([^:/\s]+):(.+)$`)

func NormalizeOrigin(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, "\r\n\t ") {
		return "", errors.New("repository URL is invalid")
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Host == "" {
			return "", errors.New("repository URL is invalid")
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "ssh" {
			return "", errors.New("repository URL has an unsupported scheme")
		}
		return canonicalOrigin(parsed.Hostname(), parsed.Path)
	}

	if match := scpStyleRemote.FindStringSubmatch(value); match != nil {
		return canonicalOrigin(match[1], match[2])
	}

	return "", errors.New("repository URL is invalid")
}

func canonicalOrigin(host, remotePath string) (string, error) {
	cleanPath := strings.TrimSuffix(strings.Trim(path.Clean("/"+remotePath), "/"), ".git")
	if strings.TrimSpace(host) == "" || cleanPath == "" || cleanPath == "." {
		return "", errors.New("repository URL is invalid")
	}
	return strings.ToLower(host) + "/" + cleanPath, nil
}
