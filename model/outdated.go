package model

type RunningImage struct {
	Namespace     string
	Pod           string
	InitContainer *string
	Container     *string
	Image         string
	PullableImage string
}

type CheckResult struct {
	IsAccessible   bool
	LatestVersion  string
	VersionsBehind int64
	CheckError     string
	Path           string
}

type CheckResultfinal struct {
	Image          string `json:"image"`
	Current        string `json:"current_tag"`
	LatestVersion  string `json:"latest_version"`
	VersionsBehind int64  `json:"versions_behind"`
	Pod            string `json:"pod"`
	Namespace      string `json:"namespace"`
	ClusterName    string `json:"clustername"`
}
