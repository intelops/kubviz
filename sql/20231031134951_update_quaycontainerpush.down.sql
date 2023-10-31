ALTER TABLE quaycontainerpush
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE quaycontainerpush
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE quaycontainerpush
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE quaycontainerpush
    SETTINGS index_granularity = PreviousIndexGranularity;