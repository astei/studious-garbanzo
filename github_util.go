package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// verifyGitHubEventSignature validates the signature sent by GitHub to remote repositories
// when a secret is provided for the webhook.
func verifyGitHubEventSignature(providedSignature string, secret string, body []byte) bool {
	if providedSignature == "" || !strings.HasPrefix(providedSignature, "sha1=") {
		return false
	}

	signature, err := hex.DecodeString(providedSignature[5:])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	computed := mac.Sum(nil)

	return hmac.Equal(computed, signature)
}

func getEventPayload(r *http.Request, body *bytes.Buffer) (io.Reader, error) {
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		v, err := url.ParseQuery(body.String())
		if err != nil {
			return nil, err
		}
		return strings.NewReader(v.Get("payload")), nil
	}
	return body, nil
}
