CREATE TABLE IF NOT EXISTS new_events (
    ClusterName String,
    Id          String,
    EventTime   DateTime('UTC'),
    OpType      String,
    Name        String,
    Namespace   String,
    Kind        String,
    Message     String,
    Reason      String,
    Host        String,
    Event       String,
    FirstTime   String,
    LastTime    String
) ENGINE = MergeTree()
ORDER BY ClusterName, EventTime
TTL EventTime + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

INSERT INTO new_events SELECT * FROM events;

RENAME TABLE events TO old_events, new_events TO events;

DROP TABLE old_events;