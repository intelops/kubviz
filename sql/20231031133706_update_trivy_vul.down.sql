ALTER TABLE trivy_vul
    ENGINE = PreviousEngine;

-- Revert the ORDER BY clause to its previous state
ALTER TABLE trivy_vul
    ORDER BY PreviousOrderBy;

-- Remove the TTL setting to disable it
ALTER TABLE trivy_vul
    CLEAR TTL;

-- Reset the index granularity to its previous value
ALTER TABLE trivy_vul
    SETTINGS index_granularity = PreviousIndexGranularity