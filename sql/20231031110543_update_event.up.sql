ALTER TABLE events
    ENGINE = MergeTree()
    ORDER BY (ClusterName, EventTime)
    TTL EventTime TO EventTime + INTERVAL 10 MINUTE;
