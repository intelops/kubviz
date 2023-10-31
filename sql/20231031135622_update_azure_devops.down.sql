ALTER TABLE azure_devops
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE azure_devops
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE azure_devops
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE azure_devops
    SETTINGS index_granularity = PreviousIndexGranularity;