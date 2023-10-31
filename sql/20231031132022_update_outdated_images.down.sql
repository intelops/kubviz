ALTER TABLE outdated_images
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE outdated_images
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE outdated_images
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE outdated_images
    SETTINGS index_granularity = PreviousIndexGranularity;