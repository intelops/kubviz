ALTER TABLE trivysbom
    ENGINE = MergeTree()
	ORDER BY (ClusterName, metadata_timestamp) 
	TTL metadata_timestamp + INTERVAL 30 DAY
	SETTINGS index_granularity = 8192;