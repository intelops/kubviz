
CREATE TABLE IF NOT EXISTS outdated_images (
	ClusterName     String,
	Namespace       String,
	Pod             String,
	CurrentImage    String,
	CurrentTag      String,
	LatestVersion   String,
	VersionsBehind  Int64,
    EventTime       DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;