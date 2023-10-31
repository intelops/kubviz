ALTER TABLE rakkess
    ENGINE = MergeTree()
    ORDER BY (ClusterName, EventTime)
    TTL EventTime + INTERVAL 10 MINUTE
    SETTINGS index_granularity = 8192;