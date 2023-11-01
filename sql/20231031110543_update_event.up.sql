ALTER TABLE events
    ENGINE = MergeTree()
    ORDER BY (ClusterName, EventTime)
    TTL EventTime + INTERVAL 10 MINUTE;
   