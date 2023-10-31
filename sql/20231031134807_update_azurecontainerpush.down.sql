ALTER TABLE azurecontainerpush
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE azurecontainerpush
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE azurecontainerpush
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE azurecontainerpush
    SETTINGS index_granularity = PreviousIndexGranularity