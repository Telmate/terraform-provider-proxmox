package sshkeys

import (
	"crypto"
	"regexp"
	"strings"
)

var regexMultipleSpaces = regexp.MustCompile(`\s+`)

func Split(rawKeys string) *[]crypto.PublicKey {
	tmpKeys := strings.Split(rawKeys, "\n")
	keys := make([]crypto.PublicKey, len(tmpKeys))
	for i, e := range tmpKeys {
		keys[i] = crypto.PublicKey(e)
	}
	return &keys
}

func String(keys *[]crypto.PublicKey) string {
	if keys != nil {
		var rawKeys string
		for _, key := range *keys {
			rawKeys += "\n" + key.(string)
		}
		if rawKeys != "" {
			return rawKeys[1:]
		}
	}
	return ""
}

func Trim(rawKeys string) string {
	return regexMultipleSpaces.ReplaceAllString(strings.TrimSpace(rawKeys), " ")
}
