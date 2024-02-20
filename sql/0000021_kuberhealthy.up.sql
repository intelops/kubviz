CREATE TABLE IF NOT EXISTS kuberhealthy (
    CurrentUUID String,
    CheckName String,
    OK UInt8,
    Errors String,
    RunDuration String,
    Namespace String,
    Node String,
    LastRun DateTime('UTC'),
    AuthoritativePod String,
    ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
    ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree()
ORDER BY ExpiryDate
TTL ExpiryDate;
