-- Create a new table with the desired storage engine and schema
CREATE TABLE new_events
ENGINE = MergeTree()
ORDER BY (ClusterName, EventTime)
TTL EventTime TO EventTime + INTERVAL 10 MINUTE AS
SELECT * FROM events;

-- You may need to update indexes, constraints, and other properties as well.

-- Copy data from the old table to the new one
INSERT INTO new_events SELECT * FROM events;

-- Drop the old table (if it's no longer needed)
DROP TABLE events;

-- Rename the new table to match the old table's name
RENAME TABLE new_events TO events;
