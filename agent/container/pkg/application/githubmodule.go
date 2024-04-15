package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Version struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Package struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PackageType  string `json:"package_type"`
	Owner        Owner  `json:"owner"`
	VersionCount int    `json:"version_count"`
	Visibility   string `json:"visibility"`
	URL          string `json:"url"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	HTMLURL      string `json:"html_url"`
}
type Owner struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GithubApiClient struct {
	Org   string
	Token string
}

func NewGithubClient(org string, token string) *GithubApiClient {
	return &GithubApiClient{Org: org, Token: token}
}

func (c *GithubApiClient) FetchPackages() ([]Package, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/packages?package_type=container", c.Org)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var packages []Package
	err = json.Unmarshal(body, &packages)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (c *GithubApiClient) FetchVersions(packageName string) ([]Version, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/packages/container/%s/versions", c.Org, packageName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Check if the version is not found before unmarshalling
	if string(body) == `{"message":"Not Found","documentation_url":"https://docs.github.com/rest"}` {
		fmt.Println("Version not found for package:", packageName)
		return nil, fmt.Errorf("Version not found for package: Please provide proper semantic version for %s image", packageName)
	}
	var versions []Version
	err = json.Unmarshal(body, &versions)
	if err != nil {
		return nil, fmt.Errorf("error occurred while unmarshalling the version %w", err)
	}

	return versions, nil
}
