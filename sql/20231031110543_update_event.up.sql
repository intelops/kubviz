-- Change the engine to MergeTree
ALTER TABLE events ENGINE = MergeTree();

-- Specify the new ORDER BY clause
ALTER TABLE events ORDER BY ClusterName, EventTime;

-- Set a TTL policy on the EventTime column
ALTER TABLE events TTL EventTime + INTERVAL 30 DAY;

-- Adjust the index granularity
ALTER TABLE events SETTINGS index_granularity = 8192;
