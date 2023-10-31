ALTER TABLE kubescore
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE kubescore
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE kubescore
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE kubescore
    SETTINGS index_granularity = PreviousIndexGranularity