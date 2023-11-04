CREATE TABLE IF NOT EXISTS rakkess (
	ClusterName String,
	Name        String,
	Create      String,
	Delete      String,
	List        String,
	Update      String,
	EventTime   DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

