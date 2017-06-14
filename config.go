package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/taskcluster/shell"
)

// Configuration is a Go representation of the contents of config.json.
type Configuration struct {
	// Where studious-garbanzo should listen.
	Listen string `json:"Listen"`

	// Security-related settings.
	Security struct {
		// Place the webhook at a custom subpath.
		CustomPath string `json:"CustomPath"`

		// A HMAC secret.
		Secret string `json:"Secret"`
	} `json:"Security"`

	// The repositories to respond to. If a repository is not listed,
	// studious-garbanzo will not do anything.
	Repositories []struct {
		// The repository to use.
		Repository string `json:"Repository"`

		// The commands to execute for this repository.
		Commands []RepoCommand `json:"Commands"`
	} `json:"Repositories"`
}

// RepoCommand represents a command to run for this repository.
type RepoCommand struct {
	// The command to execute.
	Command string `json:"Command"`

	// The arguments to use.
	Args []string `json:"Args"`

	// The current directory to use when running the command.
	Cwd string `json:"Cwd"`
}

// GetArgs gets the parsed and escaped arguments for the command.
func (rc RepoCommand) GetArgs(e github.PushEvent, repo string) []string {
	args := make([]string, len(rc.Args))
	var buf bytes.Buffer
	for i, arg := range rc.Args {
		finalArg := arg
		tmpl, err := template.New("").Parse(arg)
		if err == nil {
			if err = tmpl.Execute(&buf, e); err == nil {
				finalArg = buf.String()
			}
			buf.Truncate(0)
		}
		if err != nil {
			log.Printf("NOTICE: arg '%s' (#%d) for '%s' (%s) is not valid: %s", arg, i, rc.Command, repo, err.Error())
		}
		args[i] = shell.Escape(finalArg)
	}
	return args
}
