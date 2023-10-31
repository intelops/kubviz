CREATE TABLE IF NOT EXISTS github (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    DateTime('UTC'),
	Event        String
) engine=File(TabSeparated);