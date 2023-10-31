ALTER TABLE github
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE github
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE github
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE github
    SETTINGS index_granularity = PreviousIndexGranularity;