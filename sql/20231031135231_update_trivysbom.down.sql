ALTER TABLE trivysbom
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE trivysbom
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE trivysbom
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE trivysbom
    SETTINGS index_granularity = PreviousIndexGranularity;