CREATE TABLE IF NOT EXISTS getall_resources (
	ClusterName String,
	Namespace   String,
	Kind        String,
	Resource    String,
	Age         String,
    EventTime   DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

