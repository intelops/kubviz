package azure

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// parse errors
var (
	ErrEventNotSpecifiedToParse  = errors.New("no Event specified to parse")
	ErrInvalidHTTPMethod         = errors.New("invalid HTTP Method")
	ErrMissingGithubEventHeader  = errors.New("missing X-Azure-Event Header")
	ErrMissingHubSignatureHeader = errors.New("missing X-Hub-Signature Header")
	ErrEventNotFound             = errors.New("event not defined to be parsed")
	ErrParsingPayload            = errors.New("error parsing payload")
	ErrHMACVerificationFailed    = errors.New("HMAC verification failed")
)

// Event defines a Azure hook event type
type Event string

// Azure hook types
const (
	PushEvent                      Event = "push"
	PullRequestCreatedEvent        Event = "pull"
	PullRequestMergeAttemptedEvent Event = "merge"
	PullRequestCommentEvent        Event = "pull_comment"
)

// EventSubtype defines a Azure Hook Event subtype
type EventSubtype string

// Azure hook event subtypes
const (
	NoSubtype     EventSubtype = ""
	BranchSubtype EventSubtype = "branch"
	TagSubtype    EventSubtype = "tag"
	PullSubtype   EventSubtype = "pull"
	IssueSubtype  EventSubtype = "issues"
)

// Option is a configuration option for the webhook
type Option func(*Webhook) error

// WebhookOptions is a namespace for configuration option methods
type WebhookOptions struct{}

// Options is a namespace var for configuration options
var Options = WebhookOptions{}

// Secret registers the Azure secret
func (WebhookOptions) Secret(secret string) Option {
	return func(hook *Webhook) error {
		hook.secret = secret
		return nil
	}
}

// Webhook instance contains all methods needed to process events
type Webhook struct {
	secret string
}

// New creates and returns a WebHook instance denoted by the Provider type
func New(options ...Option) (*Webhook, error) {
	hook := new(Webhook)
	for _, opt := range options {
		if err := opt(hook); err != nil {
			return nil, errors.New("error applying option")
		}
	}
	return hook, nil
}

// Parse verifies and parses the events specified and returns the payload object or an error
func (hook Webhook) Parse(r *http.Request, events ...Event) (interface{}, error) {
	defer func() {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}()

	if len(events) == 0 {
		return nil, ErrEventNotSpecifiedToParse
	}
	if r.Method != http.MethodPost {
		return nil, ErrInvalidHTTPMethod
	}

	event := r.Header.Get("X-Azure-Event")
	if event == "" {
		return nil, ErrMissingGithubEventHeader
	}
	azureEvent := Event(event)

	var found bool
	for _, evt := range events {
		if evt == azureEvent {
			found = true
			break
		}
	}
	// event not defined to be parsed
	if !found {
		return nil, ErrEventNotFound
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, ErrParsingPayload
	}

	// If we have a Secret set, we should check the MAC
	if len(hook.secret) > 0 {
		signature := r.Header.Get("X-Hub-Signature")
		if len(signature) == 0 {
			return nil, ErrMissingHubSignatureHeader
		}
		mac := hmac.New(sha1.New, []byte(hook.secret))
		_, _ = mac.Write(payload)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature[5:]), []byte(expectedMAC)) {
			return nil, ErrHMACVerificationFailed
		}
	}

	switch azureEvent {
	case PushEvent:
		var pl PushPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case PullRequestCreatedEvent:
		var pl PullRequestCreatedPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case PullRequestMergeAttemptedEvent:
		var pl PullRequestMergeAttemptedPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case PullRequestCommentEvent:
		var pl PullRequestCommentedOnPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	default:
		return nil, fmt.Errorf("unknown event %s", azureEvent)
	}
}
