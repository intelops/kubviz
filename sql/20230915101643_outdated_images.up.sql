CREATE TABLE IF NOT EXISTS outdated_images (
	ClusterName     String,
	Namespace       String,
	Pod             String,
	CurrentImage    String,
	CurrentTag      String,
	LatestVersion   String,
	VersionsBehind  Int64,
    EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

