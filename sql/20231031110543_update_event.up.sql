CREATE TABLE IF NOT EXISTS new_events (
    ClusterName String,
    Id String,
    EventTime DateTime('UTC'),
    OpType String,
    Name String,
    Namespace String,
    Kind String,
    Message String,
    Reason String,
    Host String,
    Event String,
    FirstTime String,
    LastTime String
) ENGINE = MergeTree()
ORDER BY ClusterName
TTL EventTime + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

