CREATE TABLE IF NOT EXISTS kuberhealthy (
    CurrentUUID String,
    CheckName String,
    OK UInt8,
    Errors String,
    RunDuration String,
    Namespace String,
    Node String,
    LastRun String,
    AuthoritativePod String,
    ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree()
ORDER BY ExpiryDate
TTL ExpiryDate;
