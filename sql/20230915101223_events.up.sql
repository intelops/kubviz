CREATE TABLE IF NOT EXISTS events (
	ClusterName String,
	Id          String,
	EventTime   DateTime('UTC'),
	OpType      String,
	Name        String,
	Namespace   String,
	Kind        String,
	Message     String,
	Reason      String,
	Host        String,
	Event       String,
	FirstTime   String,
	LastTime    String,
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
