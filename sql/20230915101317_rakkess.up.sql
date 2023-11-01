CREATE TABLE IF NOT EXISTS rakkess (
	ClusterName String,
	Name        String,
	Create      String,
	Delete      String,
	List        String,
	Update      String,
	EventTime   DateTime('UTC')
) ENGINE = MergeTree()
	ORDER BY (ClusterName, EventTime) 
	TTL EventTime + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;
	
