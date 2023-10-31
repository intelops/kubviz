CREATE TABLE IF NOT EXISTS quaycontainerpush (
	name          String,
	repository    String,
	nameSpace     String,
	dockerURL     String,
	homePage      String,
	tag           String,
	Event         String,
    EventTime DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;
