CREATE TABLE IF NOT EXISTS outdated_images (
	ClusterName     String,
	Namespace       String,
	Pod             String,
	CurrentImage    String,
	CurrentTag      String,
	LatestVersion   String,
	VersionsBehind  Int64,
    EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
	ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree()
ORDER BY ExpiryDate
TTL ExpiryDate;
