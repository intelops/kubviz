package dbstatement

type DBStatement string

const AzureDevopsTable DBStatement = `
CREATE TABLE IF NOT EXISTS azure_devops (
	Author String,
	Provider String,
	CommitID String,
	CommitUrl String,
	EventType String,
	RepoName String,
	TimeStamp DateTime('UTC'),
	Event String
) ENGINE = MergeTree()
ORDER BY (ClusterName, TimeStamp) 
TTL TimeStamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
`

const InsertAzureDevops DBStatement = "INSERT INTO azure_devops ( Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

const GithubTable DBStatement = `
CREATE TABLE IF NOT EXISTS github (
	Author String,
	Provider String,
	CommitID String,
	CommitUrl String,
	EventType String,
	RepoName String,
	TimeStamp DateTime('UTC'),
	Event String
) ENGINE = MergeTree()
ORDER BY (ClusterName, TimeStamp) 
TTL TimeStamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
`

const InsertGithub DBStatement = "INSERT INTO github ( Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

const GitlabTable DBStatement = `
CREATE TABLE IF NOT EXISTS gitlab (
	Author String,
	Provider String,
	CommitID String,
	CommitUrl String,
	EventType String,
	RepoName String,
	TimeStamp DateTime('UTC'),
	Event String
) ENGINE = MergeTree()
ORDER BY (ClusterName, TimeStamp) 
TTL TimeStamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
`

const InsertGitlab DBStatement = "INSERT INTO gitlab ( Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

const BitbucketTable DBStatement = `
CREATE TABLE IF NOT EXISTS bitbucket (
	Author String,
	Provider String,
	CommitID String,
	CommitUrl String,
	EventType String,
	RepoName String,
	TimeStamp DateTime('UTC'),
	Event String
) ENGINE = MergeTree()
ORDER BY (ClusterName, TimeStamp) 
TTL TimeStamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
`

const InsertBitbucket DBStatement = "INSERT INTO bitbucket ( Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

const GiteaTable DBStatement = `
CREATE TABLE IF NOT EXISTS gitea (
	Author String,
	Provider String,
	CommitID String,
	CommitUrl String,
	EventType String,
	RepoName String,
	TimeStamp DateTime('UTC'),
	Event String
) ENGINE = MergeTree()
ORDER BY (ClusterName, TimeStamp) 
TTL TimeStamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
`
const InsertGitea DBStatement = "INSERT INTO gitea ( Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
