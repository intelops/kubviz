CREATE TABLE IF NOT EXISTS DeletedAPIs (
	ClusterName     String,
	ObjectName      String,
	Group           String,
	Kind            String,
	Version         String,
	Name            String,
	Deleted         UInt8,
	Scope           String,
	EventTime       DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;
	
