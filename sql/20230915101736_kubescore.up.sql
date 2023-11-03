CREATE TABLE IF NOT EXISTS kubescore (
	id              UUID,
	namespace       String,
	cluster_name    String,
	recommendations String,
    EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

