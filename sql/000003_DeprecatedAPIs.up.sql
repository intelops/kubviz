CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
	ClusterName     String,
	ObjectName      String,
	Description     String,
	Kind            String,
	Deprecated      UInt8,
	Scope           String,
	EventTime       DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
	ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

