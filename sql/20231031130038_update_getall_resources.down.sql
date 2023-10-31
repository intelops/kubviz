ALTER TABLE getall_resources
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE getall_resources
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE getall_resources
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE getall_resources
    SETTINGS index_granularity = PreviousIndexGranularity;