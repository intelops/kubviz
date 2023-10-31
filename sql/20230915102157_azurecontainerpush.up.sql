CREATE TABLE IF NOT EXISTS azurecontainerpush (
	RegistryURL    String,
	RepositoryName String,
	Tag            String,
	ImageName      String,
	Event          String,
	Size           Int32,
	SHAID          String,
    EventTime DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;
