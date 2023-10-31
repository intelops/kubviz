ALTER TABLE dockerhubbuild
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE dockerhubbuild
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE dockerhubbuild
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE dockerhubbuild
    SETTINGS index_granularity = PreviousIndexGranularity;