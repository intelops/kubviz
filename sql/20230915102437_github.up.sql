CREATE TABLE IF NOT EXISTS github (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    DateTime('UTC'),
	Event        String
) ENGINE = MergeTree()
	ORDER BY (ClusterName, TimeStamp) 
	TTL TimeStamp + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;
