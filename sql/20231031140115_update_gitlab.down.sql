ALTER TABLE gitlab
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE gitlab
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE gitlab
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE gitlab
    SETTINGS index_granularity = PreviousIndexGranularity;