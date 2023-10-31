ALTER TABLE trivyimage
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE trivyimage
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE trivyimage
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE trivyimage
    SETTINGS index_granularity = PreviousIndexGranularity;