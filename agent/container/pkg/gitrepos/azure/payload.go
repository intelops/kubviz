package azure

import "time"

type PushPayload struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"detailedMessage"`
	Resource struct {
		Commits []struct {
			CommitID string `json:"commitId"`
			Author   struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Committer struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"committer"`
			Comment string `json:"comment"`
			URL     string `json:"url"`
		} `json:"commits"`
		RefUpdates []struct {
			Name        string `json:"name"`
			OldObjectID string `json:"oldObjectId"`
			NewObjectID string `json:"newObjectId"`
		} `json:"refUpdates"`
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				URL            string `json:"url"`
				State          string `json:"state"`
				Visibility     string `json:"visibility"`
				LastUpdateTime string `json:"lastUpdateTime"`
			} `json:"project"`
			DefaultBranch string `json:"defaultBranch"`
			RemoteURL     string `json:"remoteUrl"`
		} `json:"repository"`
		PushedBy struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			Links       struct {
				Avatar struct {
					Href string `json:"href"`
				} `json:"avatar"`
			} `json:"_links"`
			ID         string `json:"id"`
			UniqueName string `json:"uniqueName"`
			ImageURL   string `json:"imageUrl"`
			Descriptor string `json:"descriptor"`
		} `json:"pushedBy"`
		PushID int       `json:"pushId"`
		Date   time.Time `json:"date"`
		URL    string    `json:"url"`
		Links  struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			Repository struct {
				Href string `json:"href"`
			} `json:"repository"`
			Commits struct {
				Href string `json:"href"`
			} `json:"commits"`
			Pusher struct {
				Href string `json:"href"`
			} `json:"pusher"`
			Refs struct {
				Href string `json:"href"`
			} `json:"refs"`
		} `json:"_links"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"collection"`
		Account struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"account"`
		Project struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}

type PullRequestCreatedPayload struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"detailedMessage"`
	Resource struct {
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				URL            string `json:"url"`
				State          string `json:"state"`
				Visibility     string `json:"visibility"`
				LastUpdateTime string `json:"lastUpdateTime"`
			} `json:"project"`
			DefaultBranch string `json:"defaultBranch"`
			RemoteURL     string `json:"remoteUrl"`
		} `json:"repository"`
		PullRequestID int    `json:"pullRequestId"`
		Status        string `json:"status"`
		CreatedBy     struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
		} `json:"createdBy"`
		CreationDate          time.Time `json:"creationDate"`
		Title                 string    `json:"title"`
		Description           string    `json:"description"`
		SourceRefName         string    `json:"sourceRefName"`
		TargetRefName         string    `json:"targetRefName"`
		MergeStatus           string    `json:"mergeStatus"`
		MergeID               string    `json:"mergeId"`
		LastMergeSourceCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeSourceCommit"`
		LastMergeTargetCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeTargetCommit"`
		LastMergeCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeCommit"`
		Reviewers []struct {
			ReviewerURL interface{} `json:"reviewerUrl"`
			Vote        int         `json:"vote"`
			DisplayName string      `json:"displayName"`
			URL         string      `json:"url"`
			ID          string      `json:"id"`
			UniqueName  string      `json:"uniqueName"`
			ImageURL    string      `json:"imageUrl"`
			IsContainer bool        `json:"isContainer"`
		} `json:"reviewers"`
		Commits []struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"commits"`
		URL   string `json:"url"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
			Statuses struct {
				Href string `json:"href"`
			} `json:"statuses"`
		} `json:"_links"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID string `json:"id"`
		} `json:"collection"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}

type PullRequestMergeAttemptedPayload struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"detailedMessage"`
	Resource struct {
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				URL            string `json:"url"`
				State          string `json:"state"`
				Visibility     string `json:"visibility"`
				LastUpdateTime string `json:"lastUpdateTime"`
			} `json:"project"`
			DefaultBranch string `json:"defaultBranch"`
			RemoteURL     string `json:"remoteUrl"`
		} `json:"repository"`
		PullRequestID int    `json:"pullRequestId"`
		Status        string `json:"status"`
		CreatedBy     struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
		} `json:"createdBy"`
		CreationDate          time.Time `json:"creationDate"`
		ClosedDate            time.Time `json:"closedDate"`
		Title                 string    `json:"title"`
		Description           string    `json:"description"`
		SourceRefName         string    `json:"sourceRefName"`
		TargetRefName         string    `json:"targetRefName"`
		MergeStatus           string    `json:"mergeStatus"`
		MergeID               string    `json:"mergeId"`
		LastMergeSourceCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeSourceCommit"`
		LastMergeTargetCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeTargetCommit"`
		LastMergeCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeCommit"`
		Reviewers []struct {
			ReviewerURL string `json:"reviewerUrl"`
			Vote        int    `json:"vote"`
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
			IsContainer bool   `json:"isContainer"`
		} `json:"reviewers"`
		Commits []struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"commits"`
		URL   string `json:"url"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
			Statuses struct {
				Href string `json:"href"`
			} `json:"statuses"`
		} `json:"_links"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID string `json:"id"`
		} `json:"collection"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}

type PullRequestCommentedOnPayload struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"detailedMessage"`
	Resource struct {
		ID              int `json:"id"`
		ParentCommentID int `json:"parentCommentId"`
		Author          struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
		} `json:"author"`
		Content                string    `json:"content"`
		PublishedDate          time.Time `json:"publishedDate"`
		LastUpdatedDate        time.Time `json:"lastUpdatedDate"`
		LastContentUpdatedDate time.Time `json:"lastContentUpdatedDate"`
		CommentType            string    `json:"commentType"`
		Links                  struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			Repository struct {
				Href string `json:"href"`
			} `json:"repository"`
			Threads struct {
				Href string `json:"href"`
			} `json:"threads"`
		} `json:"_links"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID string `json:"id"`
		} `json:"collection"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}
