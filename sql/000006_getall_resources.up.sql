CREATE TABLE IF NOT EXISTS getall_resources (
	ClusterName String,
	Namespace   String,
	Kind        String,
	Resource    String,
	Age         String,
    EventTime   DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}}
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

