ALTER TABLE rakkess
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE rakkess
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE rakkess
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE rakkess
    SETTINGS index_granularity = PreviousIndexGranularity;