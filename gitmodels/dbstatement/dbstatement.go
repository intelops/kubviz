package dbstatement

type DBStatement string

const AzureDevopsTable DBStatement = `
	CREATE TABLE IF NOT EXISTS azure_devops (
		RepositoryName String,
		Author String,
		Provider String,
		CommitID String,
		CommitUrl String,
		EventType String,
		RepoName String,
		TimeStamp String,
		Event String
	) engine=File(TabSeparated)
`

const InsertAzureDevops DBStatement = "INSERT INTO azure_devops (RepositoryName, Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

const GithubTable DBStatement = `
	CREATE TABLE IF NOT EXISTS github (
		RepositoryName String,
		Author String,
		Provider String,
		CommitID String,
		CommitUrl String,
		EventType String,
		RepoName String,
		TimeStamp String,
		Event String
	) engine=File(TabSeparated)
`

const InsertGithub DBStatement = "INSERT INTO github (RepositoryName, Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

const GitlabTable DBStatement = `
	CREATE TABLE IF NOT EXISTS gitlab (
		RepositoryName String,
		Author String,
		Provider String,
		CommitID String,
		CommitUrl String,
		EventType String,
		RepoName String,
		TimeStamp String,
		Event String
	) engine=File(TabSeparated)
`
const InsertGitlab DBStatement = "INSERT INTO gitlab (RepositoryName, Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

const BitbucketTable DBStatement = `
	CREATE TABLE IF NOT EXISTS bitbucket (
		RepositoryName String,
		Author String,
		Provider String,
		CommitID String,
		CommitUrl String,
		EventType String,
		RepoName String,
		TimeStamp String,
		Event String
	) engine=File(TabSeparated)
`

const InsertBitbucket DBStatement = "INSERT INTO bitbucket (RepositoryName, Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

const GiteaTable DBStatement = `
	CREATE TABLE IF NOT EXISTS gitea (
		RepositoryName String,
		Author String,
		Provider String,
		CommitID String,
		CommitUrl String,
		EventType String,
		RepoName String,
		TimeStamp String,
		Event String
	) engine=File(TabSeparated)
`

const InsertGitea DBStatement = "INSERT INTO gitea (RepositoryName, Author, Provider, CommitID, CommitUrl, EventType, RepoName, TimeStamp, Event) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
