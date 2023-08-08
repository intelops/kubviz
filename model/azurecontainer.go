package model

import "time"

type AzureContainerPushEventPayload struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    struct {
		MediaType  string `json:"mediaType"`
		Size       int32  `json:"size"`
		Digest     string `json:"digest"`
		Length     int32  `json:"length"` // Same as Size field
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"target"`
	Request struct {
		ID        string `json:"id"`
		Host      string `json:"host"`
		Method    string `json:"method"`
		UserAgent string `json:"useragent"`
	} `json:"request"`
}
