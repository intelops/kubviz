package model

import "time"

type EventKey string
type EventValue string
type GitEvent string
type GitProvider string

// Git header keys to get the events from header
var (
	GithubHeader    EventKey = "X-GitHub-Event"
	GitlabHeader    EventKey = "X-Gitlab-Event"
	BitBucketHeader EventKey = "X-Event-Key"
	GiteaHeader     EventKey = "X-Gitea-Event"
	AzureHeader     EventKey = "X-Azure_Event"
)

const (
	GithubProvider      GitProvider = "Github"
	GitlabProvider      GitProvider = "Gitlab"
	GiteaProvider       GitProvider = "Gitea"
	BitBucketProvider   GitProvider = "BitBucket"
	AzureDevopsProvider GitProvider = "AzureDevops"
)

type Event string
type Date time.Time

type BasicEvent struct {
	ID          string `json:"id"`
	EventType   Event  `json:"eventType"`
	PublisherID string `json:"publisherId"`
	Scope       string `json:"scope"`
	CreatedDate Date   `json:"createdDate"`
}
type GitCommonAttribute struct {
	Author      string
	GitProvider string
	CommitID    string
	CommitUrl   string
	EventType   string
	RepoName    string
	TimeStamp   string
	Event       string
}
