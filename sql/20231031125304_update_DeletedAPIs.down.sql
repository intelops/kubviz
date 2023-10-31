ALTER TABLE DeletedAPIs
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE DeletedAPIs
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE DeletedAPIs
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE DeletedAPIs
    SETTINGS index_granularity = PreviousIndexGranularity;