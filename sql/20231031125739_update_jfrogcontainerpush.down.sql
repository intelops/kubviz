ALTER TABLE jfrogcontainerpush
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE jfrogcontainerpush
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE jfrogcontainerpush
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE jfrogcontainerpush
    SETTINGS index_granularity = PreviousIndexGranularity