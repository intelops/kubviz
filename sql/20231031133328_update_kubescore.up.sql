CREATE TABLE IF NOT EXISTS kubescore (
	id              UUID,
	namespace       String,
	cluster_name    String,
	recommendations String,
    EventTime       DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;