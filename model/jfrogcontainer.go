package model

type JfrogContainerPushEventPayload struct {
	Domain    string `json:"domain"`
	EventType string `json:"event_type"`
	Data      struct {
		RepoKey   string `json:"repo_key"`
		Path      string `json:"path"`
		Name      string `json:"name"`
		SHA256    string `json:"sha256"`
		Size      int    `json:"size"`
		ImageName string `json:"image_name"`
		Tag       string `json:"tag"`
	} `json:"data"`
	SubscriptionKey string `json:"subscription_key"`
	JPDOrigin       string `json:"jpd_origin"`
	Source          string `json:"source"`
}
