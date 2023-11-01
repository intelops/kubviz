CREATE TABLE IF NOT EXISTS jfrogcontainerpush (
	Domain         String,
	EventType      String,
	RegistryURL    String,
	RepositoryName String,
	SHAID          String,
	Size           Int32,
	ImageName      String,
	Tag            String,
	Event          String,
    EventTime      DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;