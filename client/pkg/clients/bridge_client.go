package clients

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/go-playground/webhooks/v6/bitbucket"
	"github.com/go-playground/webhooks/v6/gitea"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/gitmodels/azuremodel"
	"github.com/intelops/kubviz/gitmodels/dbstatement"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ErrHeaderEmpty defines an error occur when header is empty in git stream
var (
	ErrHeaderEmpty = errors.New("git header is empty while subscribing from agent")
)

// GitNats specifies a Git related jetstream subjects, subject and consumer names
type GitNats string

const (
	bridgeSubjects GitNats = "GITMETRICS.*"
	bridgeSubject  GitNats = "GITMETRICS.git"
	bridgeConsumer GitNats = "Git-Consumer"
)

// SubscribeGitBridgeNats subscribes to nats jetstream and calls
// the respective funcs to insert data into clickhouse DB
func (n *NATSContext) SubscribeGitBridgeNats(conn clickhouse.DBInterface) {
	log.Printf("Creating nats consumer %s with subject: %s \n", bridgeConsumer, bridgeSubject)

	ctx := context.Background()
	tracer := otel.Tracer("git-client")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "SubscribeGitBridgeNats")
	span.SetAttributes(attribute.String("git-subscribe", "Subscribe"))
	defer span.End()

	n.stream.Subscribe(string(bridgeSubject), func(msg *nats.Msg) {
		msg.Ack()
		gitprovider := msg.Header.Get("GitProvider")
		repo := model.GitProvider(gitprovider)
		log.Printf("From SubscribeGitBridgeNats func: received payload from %v", repo)
		var eventKey model.EventKey
		switch repo {
		case model.GiteaProvider:
			eventKey = model.GiteaHeader
		case model.GithubProvider:
			eventKey = model.GithubHeader
		case model.GitlabProvider:
			eventKey = model.GitlabHeader
		case model.BitBucketProvider:
			eventKey = model.BitBucketHeader
		case model.AzureDevopsProvider:
			eventKey = model.AzureHeader
		default:
			log.Println("Unknown repo")
			return
		}
		event := msg.Header.Get(string(eventKey))
		if event == "" {
			log.Println(ErrHeaderEmpty.Error())
			return
		}
		log.Printf("event from Header %v", event)
		log.Printf("RAW DATA BEFORE UNMARSHAL: %#v", string(msg.Data))
		switch repo {
		case model.AzureDevopsProvider:
			switch event {
			case string(azuremodel.GitPushEventType):
				var pl azuremodel.GitPushEvent
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occured while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				if reflect.DeepEqual(pl.Resource, azuremodel.Resource{}) || reflect.DeepEqual(pl.Resource.PushedBy, azuremodel.PushedBy{}) {
					gca.Author = ""
				} else {
					gca.Author = pl.Resource.PushedBy.DisplayName
				}
				gca.GitProvider = string(model.AzureDevopsProvider)
				if len(pl.Resource.RefUpdates) > 0 {
					gca.CommitID = pl.Resource.RefUpdates[0].NewObjectID
				} else {
					gca.CommitID = ""
				}
				gca.CommitUrl = pl.Resource.Repository.RemoteURL + "/commit/" + pl.Resource.RefUpdates[0].NewObjectID
				gca.EventType = string(azuremodel.GitPushEventType)
				gca.RepoName = pl.Resource.Repository.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertAzureDevops)
				log.Println("Inserted AzureDevops metrics:", string(msg.Data))
				log.Println()
			case string(azuremodel.GitPullRequestMergedEventType):
				var pl azuremodel.GitPullRequestEvent
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				gca.Author = pl.Resource.CreatedBy.DisplayName
				gca.GitProvider = string(model.AzureDevopsProvider)
				gca.CommitID = pl.Resource.LastMergeCommit.CommitID
				gca.CommitUrl = pl.Resource.Repository.RemoteURL + "/commit/" + pl.Resource.LastMergeCommit.CommitID
				gca.EventType = string(azuremodel.GitPullRequestMergedEventType)
				gca.RepoName = pl.Resource.Repository.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertAzureDevops)
				log.Println("Inserted AzureDevops metrics:", string(msg.Data))
				log.Println()
			default:
				var gca model.GitCommonAttribute
				gca.GitProvider = string(model.AzureDevopsProvider)
				gca.RepoName = ""
				gca.Author = ""
				gca.CommitID = ""
				gca.CommitUrl = ""
				gca.EventType = event
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertAzureDevops)
				log.Println("Inserted GitHub metrics:", string(msg.Data))
				log.Println()
			}
		case model.GithubProvider:
			switch event {
			case string(github.PushEvent):
				var pl github.PushPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				if len(pl.Commits) > 0 {
					gca.Author = pl.Commits[0].Author.Name
				} else {
					gca.Author = ""
				}
				gca.GitProvider = string(model.GithubProvider)
				if len(pl.Commits) > 0 {
					gca.CommitID = pl.Commits[0].ID
				} else {
					gca.CommitID = pl.HeadCommit.ID
				}
				if len(pl.Commits) > 0 {
					gca.CommitUrl = pl.Commits[0].URL
				} else {
					gca.CommitUrl = pl.HeadCommit.URL
				}
				gca.EventType = string(github.PushEvent)
				gca.RepoName = pl.Repository.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGithub)
				log.Println("Inserted GitHub metrics:", string(msg.Data))
				log.Println()
			case string(github.PullRequestEvent):
				var pl github.PullRequestPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				if pl.Action == "closed" && pl.PullRequest.Merged {
					var gca model.GitCommonAttribute
					gca.Author = pl.PullRequest.User.Login
					gca.GitProvider = string(model.GithubProvider)
					if pl.PullRequest.MergeCommitSha != nil {
						gca.CommitID = *pl.PullRequest.MergeCommitSha
					} else {
						gca.CommitID = ""
					}
					gca.CommitUrl = pl.PullRequest.HTMLURL
					gca.EventType = string(github.PullRequestEvent)
					gca.RepoName = pl.Repository.Name
					gca.TimeStamp = time.Now().UTC()
					gca.Event = string(msg.Data)
					conn.InsertGitCommon(gca, dbstatement.InsertGithub)
					log.Println("Inserted GitHub metrics:", string(msg.Data))
					log.Println()
				}
			default:
				var gca model.GitCommonAttribute
				gca.RepoName = ""
				gca.Author = ""
				gca.GitProvider = string(model.GithubProvider)
				gca.CommitID = ""
				gca.CommitUrl = ""
				gca.EventType = event
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGithub)
				log.Println("Inserted GitHub metrics:", string(msg.Data))
				log.Println()
			}
		case model.GiteaProvider:
			switch event {
			case string(gitea.PushEvent):
				var pl gitea.PushPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				if len(pl.Commits) > 0 {
					gca.Author = pl.Commits[0].Author.Name
				} else {
					gca.Author = ""
				}
				gca.GitProvider = string(model.GiteaProvider)
				gca.CommitID = pl.After
				gca.CommitUrl = pl.CompareURL
				gca.EventType = string(gitea.PushEvent)
				gca.RepoName = pl.Repo.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGitea)
				log.Println("Inserted Gitea metrics:", string(msg.Data))
				log.Println()
			case string(gitea.PullRequestEvent):
				var pl gitea.PullRequestPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				if pl.Action == "closed" {
					var gca model.GitCommonAttribute
					if pl.Sender != nil {
						gca.Author = pl.Sender.FullName
					} else {
						gca.Author = ""
					}
					gca.GitProvider = string(model.GiteaProvider)
					if pl.PullRequest != nil {
						gca.CommitID = *pl.PullRequest.MergedCommitID
						gca.CommitUrl = pl.PullRequest.HTMLURL
					} else {
						gca.CommitID = ""
						gca.CommitUrl = ""
					}

					gca.EventType = string(gitea.PullRequestEvent)
					if pl.Repository != nil {
						gca.RepoName = pl.Repository.Name
					} else {
						gca.RepoName = ""
					}
					gca.TimeStamp = time.Now().UTC()
					gca.Event = string(msg.Data)
					conn.InsertGitCommon(gca, dbstatement.InsertGitea)
					log.Println("Inserted Gitea metrics:", string(msg.Data))
					log.Println()
				}
			default:
				var gca model.GitCommonAttribute
				gca.GitProvider = string(model.GiteaProvider)
				gca.CommitID = ""
				gca.CommitUrl = ""
				gca.EventType = event
				gca.TimeStamp = time.Now().UTC()
				gca.RepoName = ""
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGitea)
				log.Println("Inserted Gitea metrics:", string(msg.Data))
				log.Println()
			}
		case model.GitlabProvider:
			switch event {
			case string(gitlab.PushEvents):
				var pl gitlab.PushEventPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				if len(pl.Commits) > 0 {
					gca.Author = pl.Commits[0].Author.Name
				} else {
					gca.Author = ""
				}
				gca.GitProvider = string(model.GitlabProvider)
				gca.CommitID = pl.After
				gca.CommitUrl = pl.Project.WebURL + "/commit/" + pl.After
				gca.EventType = string(gitlab.PushEvents)
				gca.RepoName = pl.Project.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGitlab)
				log.Println("Inserted Gitlab metrics:", string(msg.Data))
				log.Println()
			case string(gitlab.MergeRequestEvents):
				var pl gitlab.MergeRequestEventPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				if pl.ObjectAttributes.State == "merged" {
					var gca model.GitCommonAttribute
					gca.Author = ""
					gca.GitProvider = string(model.GitlabProvider)
					gca.CommitID = pl.ObjectAttributes.LastCommit.ID
					gca.CommitUrl = pl.ObjectAttributes.URL
					gca.EventType = string(gitlab.MergeRequestEvents)
					gca.RepoName = pl.Project.Name
					gca.TimeStamp = time.Now().UTC()
					gca.Event = string(msg.Data)
					conn.InsertGitCommon(gca, dbstatement.InsertGitlab)
					log.Println("Inserted Gitlab metrics:", string(msg.Data))
					log.Println()
				}
			default:
				var gca model.GitCommonAttribute
				gca.Author = ""
				gca.GitProvider = string(model.GitlabProvider)
				gca.CommitID = ""
				gca.CommitUrl = ""
				gca.EventType = event
				gca.TimeStamp = time.Now().UTC()
				gca.RepoName = ""
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertGitlab)
				log.Println("Inserted Gitlab metrics:", string(msg.Data))
				log.Println()
			}
		case model.BitBucketProvider:
			switch event {
			case string(bitbucket.RepoPushEvent):
				var pl bitbucket.RepoPushPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				gca.Author = pl.Actor.DisplayName
				gca.GitProvider = string(model.BitBucketProvider)
				if len(pl.Push.Changes) > 0 {
					gca.CommitID = pl.Push.Changes[0].New.Target.Hash
					gca.CommitUrl = pl.Repository.Links.HTML.Href + "/commits/" + pl.Push.Changes[0].New.Target.Hash
				} else {
					gca.CommitID = ""
					gca.CommitUrl = ""
				}
				gca.EventType = string(bitbucket.RepoPushEvent)
				gca.RepoName = pl.Repository.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertBitbucket)
				log.Println("Inserted BitBucket metrics:", string(msg.Data))
				log.Println()
			case string(bitbucket.PullRequestMergedEvent):
				var pl bitbucket.PullRequestMergedPayload
				err := json.Unmarshal([]byte(msg.Data), &pl)
				if err != nil {
					log.Println("error occurred while unmarshal the payload Error:", err.Error())
					return
				}
				var gca model.GitCommonAttribute
				gca.Author = pl.PullRequest.Author.DisplayName
				gca.GitProvider = string(model.BitBucketProvider)
				gca.CommitID = pl.PullRequest.Destination.Commit.Hash
				gca.CommitUrl = pl.PullRequest.Links.HTML.Href
				gca.EventType = string(bitbucket.PullRequestMergedEvent)
				gca.RepoName = pl.Repository.Name
				gca.TimeStamp = time.Now().UTC()
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertBitbucket)
				log.Println("Inserted BitBucket metrics:", string(msg.Data))
				log.Println()
			default:
				var gca model.GitCommonAttribute
				gca.Author = ""
				gca.GitProvider = string(model.BitBucketProvider)
				gca.CommitID = ""
				gca.CommitUrl = ""
				gca.EventType = event
				gca.TimeStamp = time.Now().UTC()
				gca.RepoName = ""
				gca.Event = string(msg.Data)
				conn.InsertGitCommon(gca, dbstatement.InsertBitbucket)
				log.Println("Inserted BitBucket metrics:", string(msg.Data))
				log.Println()
			}
		}
	}, nats.Durable(string(bridgeConsumer)), nats.ManualAck())
}
