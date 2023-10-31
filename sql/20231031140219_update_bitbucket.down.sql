ALTER TABLE bitbucket
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE bitbucket
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE bitbucket
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE bitbucket
    SETTINGS index_granularity = PreviousIndexGranularity