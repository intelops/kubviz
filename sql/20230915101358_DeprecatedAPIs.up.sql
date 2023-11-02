CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
	ClusterName     String,
	ObjectName      String,
	Description     String,
	Kind            String,
	Deprecated      UInt8,
	Scope           String,
	EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 6 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

