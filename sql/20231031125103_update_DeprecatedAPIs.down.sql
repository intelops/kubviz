ALTER TABLE DeprecatedAPIs
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE DeprecatedAPIs
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE DeprecatedAPIs
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE DeprecatedAPIs
    SETTINGS index_granularity = PreviousIndexGranularity;