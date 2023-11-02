CREATE TABLE IF NOT EXISTS gitea (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    DateTime('UTC'),
	Event        String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL 6 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
