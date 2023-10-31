ALTER TABLE gitea
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE gitea
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE gitea
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE gitea
    SETTINGS index_granularity = PreviousIndexGranularity