package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
)

var (
	ErrMissingGithubEventHeader = errors.New("missing X-GitHub-Event Header")
)

func (ah *APIHandler) PostEventDockerGithub(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		log.Printf("%v", ErrMissingGithubEventHeader)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		log.Printf("%v: %v", ErrReadingBody, err)
		return
	}

	log.Printf("Received docker event from github artifactory: %v", string(payload))
	err = ah.conn.Publish(payload, "Github_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		return
	}
}
