ALTER TABLE trivyimage
    ENGINE = MergeTree()
    ORDER BY (ClusterName, vul_last_modified_date) 
    TTL vul_last_modified_date + INTERVAL 30 DAY
    SETTINGS index_granularity = 8192;
