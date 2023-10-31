ALTER TABLE trivy_misconfig
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE trivy_misconfig
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE trivy_misconfig
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE trivy_misconfig
    SETTINGS index_granularity = PreviousIndexGranularity