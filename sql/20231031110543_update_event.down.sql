
ALTER TABLE events
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE events
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE events
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE events
    SETTINGS index_granularity = PreviousIndexGranularity;
