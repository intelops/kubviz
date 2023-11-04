CREATE TABLE IF NOT EXISTS DeletedAPIs (
	ClusterName     String,
	ObjectName      String,
	Group           String,
	Kind            String,
	Version         String,
	Name            String,
	Deleted         UInt8,
	Scope           String,
	EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
