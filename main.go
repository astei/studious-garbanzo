package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

	"io"

	"github.com/google/go-github/github"
)

var config Configuration

const (
	maxPayloadSize = 5 * 1024 * 1024 // 5MB
)

// onGitHubPush is the main HTTP handler function.
func onGitHubPush(w http.ResponseWriter, r *http.Request) {
	// smoke test
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read and buffer the whole payload, to a limit of 5MB
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)
	defer r.Body.Close()
	var body bytes.Buffer
	if _, err := body.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if config.Security.Secret != "" {
		// Verify the signature
		if !verifyGitHubEventSignature(r.Header.Get("X-Hub-Signature"), config.Security.Secret, body.Bytes()) {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid signature")
			return
		}
	}

	// check for ping (or any other events)
	if r.Header.Get("X-GitHub-Event") != "push" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var e github.PushEvent
	jsonBody, err := getEventPayload(r, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "invalid payload sent")
		return
	}
	if err := json.NewDecoder(jsonBody).Decode(&e); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "invalid JSON")
		return
	}

	for _, repoConfig := range config.Repositories {
		if repoConfig.Repository == *e.Repo.FullName {
			// Bake and execute!
			for _, cmdConfig := range repoConfig.Commands {
				// Now execute the commands
				ex := exec.Command(cmdConfig.Command, cmdConfig.GetArgs(e, repoConfig.Repository)...)
				if err := ex.Run(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "couldn't run command: "+err.Error())
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// oh well
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "no matching repository found")
}

func main() {
	cf, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("unable to open config.json: %s", err)
	}

	if err = json.NewDecoder(cf).Decode(&config); err != nil {
		cf.Close()
		log.Fatalf("unable to read config.json: %s", err)
	}

	cf.Close()
	http.HandleFunc("/"+config.Security.CustomPath, onGitHubPush)
	http.ListenAndServe(config.Listen, nil)
}
