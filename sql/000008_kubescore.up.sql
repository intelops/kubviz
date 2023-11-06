CREATE TABLE IF NOT EXISTS kubescore (
	id              UUID,
	namespace       String,
	cluster_name    String,
	recommendations String,
    EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

