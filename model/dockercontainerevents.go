package model

type BuildPayload struct {
	CallbackURL string `json:"callback_url"`
	PushData    struct {
		Images   []string `json:"images"`
		PushedAt float32  `json:"pushed_at"`
		Pusher   string   `json:"pusher"`
		Tag      string   `json:"tag"`
	} `json:"push_data"`
	Repository struct {
		CommentCount    int     `json:"comment_count"`
		DateCreated     float32 `json:"date_created"`
		Description     string  `json:"description"`
		Dockerfile      string  `json:"dockerfile"`
		FullDescription string  `json:"full_description"`
		IsOfficial      bool    `json:"is_official"`
		IsPrivate       bool    `json:"is_private"`
		IsTrusted       bool    `json:"is_trusted"`
		Name            string  `json:"name"`
		Namespace       string  `json:"namespace"`
		Owner           string  `json:"owner"`
		RepoName        string  `json:"repo_name"`
		RepoURL         string  `json:"repo_url"`
		StarCount       int     `json:"star_count"`
		Status          string  `json:"status"`
	} `json:"repository"`
}

type DockerHubBuild struct {
	PushedBy       string
	ImageTag       string
	RepositoryName string
	DateCreated    string
	Owner          string
	Event          string
}
